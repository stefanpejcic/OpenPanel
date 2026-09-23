package docker

import (
	"sync"
	"testing"
	"time"
)

type recorded struct {
	line   string
	edited bool
}

// runRecorder feeds keystrokes and fake PTY echo in order, then waits for the delayed echo checks
func runRecorder(t *testing.T, steps ...[2]string) []recorded {
	t.Helper()
	old := echoWaitDelay
	echoWaitDelay = 10 * time.Millisecond
	t.Cleanup(func() { echoWaitDelay = old })

	var mu sync.Mutex
	var got []recorded
	out := &terminalOutput{}
	rec := newCommandRecorder(out, func(line string, edited bool) {
		mu.Lock()
		got = append(got, recorded{line, edited})
		mu.Unlock()
	})
	for _, s := range steps {
		if s[0] == "out" {
			out.Write([]byte(s[1]))
		} else {
			rec.Feed([]byte(s[1]))
		}
	}
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	return got
}

func in(s string) [2]string  { return [2]string{"in", s} }
func out(s string) [2]string { return [2]string{"out", s} }

func TestCommandRecorderLogsEchoedCommand(t *testing.T) {
	got := runRecorder(t, out("# "), in("ls -la"), out("ls -la"), in("\r"))
	if len(got) != 1 || got[0].line != "ls -la" || got[0].edited {
		t.Fatalf("got %+v", got)
	}
}

func TestCommandRecorderSkipsUnechoedInput(t *testing.T) {
	got := runRecorder(t, out("Enter password: "), in("hunter2\r"))
	if len(got) != 0 {
		t.Fatalf("password was logged: %+v", got)
	}
}

func TestCommandRecorderBackspace(t *testing.T) {
	got := runRecorder(t, in("lss"), out("lss"), in("\x7f"), out("\b\x1b[K"), in(" /tmp"), out(" /tmp"), in("\r"))
	if len(got) != 1 || got[0].line != "ls /tmp" {
		t.Fatalf("got %+v", got)
	}
}

func TestCommandRecorderHistoryRecall(t *testing.T) {
	got := runRecorder(t, in("\x1b[A"), out("ls"), in("\r"))
	if len(got) != 1 || got[0].line != "" || !got[0].edited {
		t.Fatalf("got %+v", got)
	}
}

func TestCommandRecorderCtrlCDiscardsLine(t *testing.T) {
	got := runRecorder(t, in("rm -rf x"), out("rm -rf x"), in("\x03"), out("^C\r\n# "), in("\r"))
	if len(got) != 0 {
		t.Fatalf("got %+v", got)
	}
}

func TestCommandRecorderBracketedPasteMultiLine(t *testing.T) {
	got := runRecorder(t, in("\x1b[200~echo a\recho b\r\x1b[201~"), out("echo a\r\na\r\n# echo b\r\nb\r\n"))
	if len(got) != 2 || got[0].line != "echo a" || got[1].line != "echo b" || got[0].edited {
		t.Fatalf("got %+v", got)
	}
}

func TestTerminalCommandLogMessage(t *testing.T) {
	if m := terminalCommandLogMessage("php", "ls", false); m != "executed command in terminal for service php: ls" {
		t.Fatal(m)
	}
	if m := terminalCommandLogMessage("php", "", true); m != "executed a command from shell history in terminal for service php" {
		t.Fatal(m)
	}
}
