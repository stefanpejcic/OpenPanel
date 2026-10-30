package diskusage

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

// real `find . -xdev -type f -size +1M -printf '%s\t%P\n'` lines from the test VM's home
const sampleFindOutput = "126998165\tdocker-data/volumes/testinguser_html_data/_data/matomo.tests.openpanel.org/misc/DBIP-City.mmdb\n" +
	"100663296\tdocker-data/volumes/testinguser_mysql_data/_data/ib_logfile0\n" +
	"32961410\tdocker-data/volumes/testinguser_html_data/_data/nextcloud.tests.openpanel.org/apps/password_policy/lists/list-8.php\n" +
	"1048577\tbackup_files/small.tar\n" +
	"not a line\n" +
	"x\tbad size\n"

func TestParseLargestFiles(t *testing.T) {
	files := parseLargestFiles(sampleFindOutput, "/disk-usage/", 3)
	if len(files) != 3 {
		t.Fatalf("got %d files, want 3: %+v", len(files), files)
	}
	want := []struct {
		path, size, href string
		pct              int
	}{
		{"docker-data/volumes/testinguser_html_data/_data/matomo.tests.openpanel.org/misc/DBIP-City.mmdb", "121 MB", "/files/matomo.tests.openpanel.org/misc", 100},
		{"docker-data/volumes/testinguser_mysql_data/_data/ib_logfile0", "96.0 MB", "", 79},
		{"docker-data/volumes/testinguser_html_data/_data/nextcloud.tests.openpanel.org/apps/password_policy/lists/list-8.php", "31.4 MB", "/files/nextcloud.tests.openpanel.org/apps/password_policy/lists", 25},
	}
	for i, w := range want {
		f := files[i]
		if f.Path != w.path || f.Size != w.size || f.FileManagerHref != w.href || f.Pct != w.pct {
			t.Errorf("file %d = %+v, want %+v", i, f, w)
		}
	}

	// inside the docroot volume, a file at its top opens the File Manager root
	files = parseLargestFiles("2097152\tindex.zip\n", "/disk-usage/docker-data/volumes/testinguser_html_data/_data/", 20)
	if len(files) != 1 || files[0].FileManagerHref != "/files/" || files[0].Size != "2.0 MB" {
		t.Errorf("docroot file = %+v", files)
	}
	if got := parseLargestFiles("", "/disk-usage/", 20); len(got) != 0 {
		t.Errorf("empty output = %+v", got)
	}
}

func TestImageStoragePrune(t *testing.T) {
	cases := map[string]string{".": "./docker-data/overlay*", "docker-data": "./overlay*", "docker-data/volumes": "", "files": ""}
	for rel, want := range cases {
		if got := imageStoragePrune(rel); got != want {
			t.Errorf("imageStoragePrune(%q) = %q, want %q", rel, got, want)
		}
	}
}

func TestBuildDiskKPIs(t *testing.T) {
	rows := parseDuOutput("9.2G\tdocker-data\n868K\tphp.ini\n0\timapsync", "/disk-usage/")
	quota := &dashboard.DiskInodes{DiskUsed: 5 * 1024 * 1024, DiskSoft: 10 * 1024 * 1024, InodesUsed: 312480, InodesSoft: 1000000}

	k := buildDiskKPIs(quota, rows, true, 1288490188, true)
	if k.Disk == nil || k.Disk.Value != "5.0" || k.Disk.Unit != "GB" || k.Disk.Limit != "10.0 GB" || k.Disk.Pct != 50 || k.Disk.FromListing {
		t.Errorf("disk tile = %+v", k.Disk)
	}
	if k.Inodes == nil || k.Inodes.Value != "312,480" || k.Inodes.Limit != "1,000,000" || k.Inodes.Pct != 31 {
		t.Errorf("inodes tile = %+v", k.Inodes)
	}
	if k.LargestFolder == nil || k.LargestFolder.Directory != "docker-data" || k.LargestSize != "9.2 GB" {
		t.Errorf("largest folder = %+v %q", k.LargestFolder, k.LargestSize)
	}
	if k.Trash == nil || k.Trash.Value != "1.2" || k.Trash.Unit != "GB" || k.TrashBytes != 1288490188 {
		t.Errorf("trash tile = %+v", k.Trash)
	}

	// no quota report: disk falls back to the home listing, inodes left out, trash failure left out
	k = buildDiskKPIs(nil, rows, true, 0, false)
	if k.Disk == nil || !k.Disk.FromListing || k.Disk.Unit != "GB" || k.Inodes != nil || k.Trash != nil {
		t.Errorf("fallback tiles = %+v %+v %+v", k.Disk, k.Inodes, k.Trash)
	}
	if k = buildDiskKPIs(nil, rows, false, 0, true); k.Disk != nil {
		t.Errorf("subfolder without quota shouldn't show a disk tile: %+v", k.Disk)
	}
	if k = buildDiskKPIs(&dashboard.DiskInodes{DiskUsed: 2048}, nil, true, 0, true); k.Disk.Limit != "" || k.Inodes.Limit != "" || k.LargestFolder != nil {
		t.Errorf("unlimited quota = %+v %+v", k.Disk, k.Inodes)
	}
}

func TestGroupThousands(t *testing.T) {
	for n, want := range map[int64]string{0: "0", 999: "999", 1000: "1,000", 312480: "312,480", 1000000: "1,000,000", -4500: "-4,500"} {
		if got := groupThousands(n); got != want {
			t.Errorf("groupThousands(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestRenderDiskUsageKPIs(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	layout := baseLayout(mgr, "/disk-usage/")
	layout.UserAllowed["trash"] = true
	rows := parseDuOutput("9.2G\tdocker-data\n868K\tphp.ini", "/disk-usage/")
	data := DiskUsagePageData{LayoutData: layout, Rows: rows, KPIs: buildDiskKPIs(&dashboard.DiskInodes{DiskUsed: 5 << 20, DiskSoft: 10 << 20, InodesUsed: 312480, InodesSoft: 1000000}, rows, true, 2048, true)}
	data.KPIs.TrashHref = "/files.trash"
	w := httptest.NewRecorder()
	if err := diskUsagePage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"of 10.0 GB", "312,480", "of 1,000,000", "Largest folder", "9.2 GB", `href="/files.trash"`, "Empty it to free space", "Largest files", "largestFiles()", "?largest=1"} {
		if !strings.Contains(body, want) {
			t.Errorf("disk usage page missing %q", want)
		}
	}
}
