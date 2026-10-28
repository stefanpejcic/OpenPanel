package emails

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// admin opt-in: imapsync skips messages over the limit, the fetch script downloads them in the account's container, the panel saves them with doveadm

// oversizeCap is how big a message migrations may copy past the mailserver's limit, 0 when the admin hasn't turned it on
func oversizeCap(a *appctx.App) int64 {
	return int64(migrateConfigInt(a, "imapsync_oversize_mb", 0)) << 20
}

// oversizeFetchScript runs imapsync with the given args, then downloads the messages it skipped with --maxsize into <job>.big/, args are the cap in bytes followed by imapsync's own arguments
const oversizeFetchScript = `
use strict; use warnings;
use Mail::IMAPClient; use JSON::PP; use IO::Handle;
my ($cap, @args) = @ARGV;
my %o;
for (@args) {
	if (/^--(host1|port1|user1|passfile1|logdir|logfile)=(.*)\z/s) { $o{$1} = $2 }
	elsif ($_ eq '--ssl1') { $o{ssl} = 1 }
	elsif ($_ eq '--tls1') { $o{tls} = 1 }
}
# read now, the panel removes the passfiles once imapsync has logged in
open(my $pf, '<', $o{passfile1}) or die "passfile1: $!";
my $pass = <$pf> // ''; close $pf; chomp $pass;

my $stop = 0;
my $pid = fork() // die "fork: $!";
if (!$pid) { exec('imapsync', @args) or exit 127 }
# podman stop signals pid 1, which is this script
$SIG{$_} = sub { $stop = 1; kill $_[0], $pid } for qw(INT TERM);
waitpid($pid, 0);
my $rc = ($? & 127) ? 128 + ($? & 127) : $? >> 8;
exit $rc if $stop;

my @big;
open(my $lf, '<', "$o{logdir}/$o{logfile}") or exit $rc;
while (<$lf>) { push @big, [$1, $2, $3] if m{^msg (.+)/(\d+) skipped \((\d+) exceeds maxsize limit} }
close $lf;
exit $rc unless @big;

(my $dir = "$o{logdir}/$o{logfile}") =~ s/\.log\z/.big/;
mkdir $dir;
open(my $mf, '>', "$dir/manifest.jsonl") or exit $rc;
$mf->autoflush(1);
my ($imap, $tried, $cur, $n) = (undef, 0, '', 0);
for my $m (@big) {
	last if $stop;
	my ($folder, $uid, $size) = @$m;
	my %e = (folder => $folder, uid => $uid + 0, size => $size + 0);
	if ($size > $cap) {
		$e{skip} = 'cap';
	} else {
		$imap = Mail::IMAPClient->new(Server => $o{host1}, Port => $o{port1}, User => $o{user1}, Password => $pass,
			Uid => 1, Peek => 1, Timeout => 300,
			($o{ssl} ? (Ssl => [SSL_verify_mode => 0]) : ()), ($o{tls} ? (Starttls => [SSL_verify_mode => 0]) : ())) unless $tried++;
		if (!$imap) { $e{error} = 'login to the source server failed' }
		elsif ($cur ne $folder && !$imap->examine($folder)) { $e{error} = 'could not open the folder' }
		else {
			$cur = $folder; $n++;
			if ($imap->message_to_file("$dir/$n.eml", $uid)) { $e{file} = "$n.eml" }
			else { $e{error} = 'download failed'; unlink "$dir/$n.eml" }
		}
	}
	print $mf encode_json(\%e), "\n";
}
$imap->logout if $imap;
exit $rc;
`

// oversizeEntry is one line of the manifest the fetch script writes
type oversizeEntry struct {
	Folder string `json:"folder"`
	UID    int64  `json:"uid"`
	Size   int64  `json:"size"`
	File   string `json:"file"`
	Skip   string `json:"skip"`
	Error  string `json:"error"`
}

// oversizeImport is a job whose large messages still have to be saved, from podman run until the import is done
type oversizeImport struct {
	id        string
	cancelled atomic.Bool
}

// userContext -> *oversizeImport
var oversizeImports sync.Map

// importingJob returns the import userContext has after its imapsync container, nil if none
func importingJob(userContext string) *oversizeImport {
	v, _ := oversizeImports.Load(userContext)
	imp, _ := v.(*oversizeImport)
	return imp
}

var (
	oversizeFileRE  = regexp.MustCompile(`^\d{1,7}\.eml$`)
	migrateFolderRE = regexp.MustCompile(`^Folder\s+\d+/\d+ \[(.*)\]\s+-> \[(.*)\]\s*$`)
)

func oversizeDir(id string) string { return id + ".big" }

// watchOversize waits for the job's imapsync container to exit, then saves the messages it downloaded
func watchOversize(a *appctx.App, userID int, username, userContext string, imp *oversizeImport, deadline time.Time) {
	defer oversizeImports.Delete(userContext)
	for time.Now().Before(deadline) {
		time.Sleep(5 * time.Second)
		if running, id := containerState(context.Background(), userContext, migrateContainer); running && id == imp.id {
			continue
		}
		ctx := context.Background()
		importOversize(ctx, userContext, imp, oversizeCap(a), func(dest string) bool { return ownsMailbox(ctx, a, userID, username, dest) })
		return
	}
}

// importOversize saves the downloaded messages with the size limit off and records the counts in the job file
func importOversize(ctx context.Context, userContext string, imp *oversizeImport, maxBytes int64, owns func(string) bool) {
	root, uid, gid, err := openJobDir(userContext)
	if err != nil {
		return
	}
	defer root.Close()
	defer func() { _ = root.RemoveAll(oversizeDir(imp.id)) }()
	job, err := readJob(root, imp.id)
	if err != nil || job.Cancelled {
		return
	}
	entries := readOversizeManifest(root, imp.id)

	var notes []string
	note := func(format string, args ...any) {
		// the log keeps imapsync's exit line near the end, so only the first few problems are listed
		if len(notes) < 50 {
			notes = append(notes, fmt.Sprintf(format, args...))
		}
	}
	// the job file sits in the account's home, so the destination is checked again before saving anything as root
	var folders map[string]bool
	_, setting := mailserverSizeLimit(ctx)
	blocked := ""
	switch {
	case len(entries) == 0:
	case !owns(job.Destination):
		blocked = "the destination is not one of this account's email accounts"
	case setting == "":
		blocked = "could not read the mailserver's size limit"
	default:
		if folders = destFolders(ctx, job.Destination); folders == nil {
			blocked = "could not list the destination account's folders"
		}
	}
	mapping := folderMapping(root, imp.id)

	for _, e := range entries {
		if imp.cancelled.Load() {
			break
		}
		src := e.Folder + "/" + strconv.FormatInt(e.UID, 10)
		switch {
		case e.Skip == "cap" || e.Size > maxBytes:
			job.BigOverCap++
			note("Large message %s (%d bytes) not copied, it is over the %d MB allowed", src, e.Size, maxBytes>>20)
		case e.File == "":
			job.BigFailed++
			note("Large message %s not copied: %s", src, e.Error)
		case blocked != "":
			job.BigFailed++
			note("Large message %s not copied: %s", src, blocked)
		default:
			if err := saveOversize(ctx, root, imp.id, e, job.Destination, mapping, folders, setting, maxBytes); err != nil {
				job.BigFailed++
				note("Large message %s not copied: %v", src, err)
			} else {
				job.BigCopied++
			}
		}
	}

	if len(entries) > 0 {
		lines := append([]string{fmt.Sprintf("Messages over the mailserver's size limit: %d copied, %d over %d MB, %d failed", job.BigCopied, job.BigOverCap, maxBytes>>20, job.BigFailed)}, notes...)
		appendJobLog(root, imp.id, lines)
	}
	job.Cancelled = job.Cancelled || imp.cancelled.Load()
	job.BigDone = true
	_ = writeJob(root, job, uid, gid)
}

// saveOversize stores one downloaded message in the destination folder imapsync mapped it to, unless it's already there
func saveOversize(ctx context.Context, root *os.Root, id string, e oversizeEntry, dest string, mapping map[string]string, folders map[string]bool, setting string, maxBytes int64) error {
	if !oversizeFileRE.MatchString(e.File) {
		return fmt.Errorf("bad file name")
	}
	dstRaw, ok := mapping[e.Folder]
	if !ok {
		dstRaw = e.Folder
	}
	folder, ok := decodeMUTF7(dstRaw)
	if !ok || !folders[folder] {
		return fmt.Errorf("folder %q does not exist in the destination account", dstRaw)
	}
	f, err := root.Open(oversizeDir(id) + "/" + e.File)
	if err != nil {
		return fmt.Errorf("download missing")
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return fmt.Errorf("download is over the %d MB allowed", maxBytes>>20)
	}

	if mid := messageID(f); mid != "" {
		sctx, cancel := context.WithTimeout(ctx, time.Minute)
		out, err := runMailserver(sctx, nil, "doveadm", "search", "-u", dest, "mailbox", folder, "header", "message-id", mid)
		cancel()
		if err == nil && strings.TrimSpace(string(out)) != "" {
			return nil
		}
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	sctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	if _, err := runMailserver(sctx, f, "doveadm", "-o", setting+"=0", "save", "-u", dest, "-m", folder); err != nil {
		return err
	}
	return nil
}

// readOversizeManifest reads what the fetch script downloaded, nil when it didn't run
func readOversizeManifest(root *os.Root, id string) []oversizeEntry {
	f, err := root.Open(oversizeDir(id) + "/manifest.jsonl")
	if err != nil {
		return nil
	}
	defer f.Close()
	var entries []oversizeEntry
	dec := json.NewDecoder(io.LimitReader(f, 16<<20))
	for {
		var e oversizeEntry
		if err := dec.Decode(&e); err != nil {
			return entries
		}
		entries = append(entries, e)
	}
}

// folderMapping reads imapsync's "Folder n/m [src] -> [dst]" lines from the job's log
func folderMapping(root *os.Root, id string) map[string]string {
	mapping := map[string]string{}
	f, err := root.Open(id + ".log")
	if err != nil {
		return mapping
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		if m := migrateFolderRE.FindStringSubmatch(sc.Text()); m != nil {
			mapping[m[1]] = m[2]
		}
	}
	return mapping
}

// destFolders lists the destination account's folders as doveadm names them, nil if it can't
func destFolders(ctx context.Context, dest string) map[string]bool {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	out, err := runMailserver(ctx, nil, "doveadm", "mailbox", "list", "-u", dest)
	if err != nil {
		return nil
	}
	folders := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimRight(line, "\r"); line != "" {
			folders[line] = true
		}
	}
	return folders
}

// messageID returns the Message-ID header of a message file, "" if it has none
func messageID(r io.Reader) string {
	h, _ := textproto.NewReader(bufio.NewReader(io.LimitReader(r, 256*1024))).ReadMIMEHeader()
	return strings.TrimSpace(h.Get("Message-Id"))
}

// appendJobLog adds lines to the end of a finished job's log
func appendJobLog(root *os.Root, id string, lines []string) {
	f, err := root.OpenFile(id+".log", os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = io.WriteString(f, strings.Join(lines, "\n")+"\n")
}

// decodeMUTF7 turns an IMAP modified UTF-7 folder name, as imapsync logs it, into the UTF-8 name doveadm wants
func decodeMUTF7(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '&' {
			b.WriteByte(s[i])
			continue
		}
		end := strings.IndexByte(s[i+1:], '-')
		if end < 0 {
			return "", false
		}
		enc := s[i+1 : i+1+end]
		i += end + 1
		if enc == "" {
			b.WriteByte('&')
			continue
		}
		raw, err := base64.RawStdEncoding.DecodeString(strings.ReplaceAll(enc, ",", "/"))
		if err != nil || len(raw)%2 != 0 {
			return "", false
		}
		units := make([]uint16, len(raw)/2)
		for k := range units {
			units[k] = uint16(raw[2*k])<<8 | uint16(raw[2*k+1])
		}
		b.WriteString(string(utf16.Decode(units)))
	}
	return b.String(), true
}
