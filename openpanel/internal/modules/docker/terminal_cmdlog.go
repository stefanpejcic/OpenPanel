package docker

import (
	"bytes"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	maxLoggedCommandLen = 1000
	outputLogWindow     = 64 * 1024
)

var echoWaitDelay = time.Second

// terminalOutput keeps the tail of what the PTY printed, with absolute offsets so a typed line can check what got echoed since it started
type terminalOutput struct {
	mu    sync.Mutex
	buf   []byte
	total int64
}

func (o *terminalOutput) Write(p []byte) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.buf = append(o.buf, p...)
	if len(o.buf) > outputLogWindow {
		o.buf = append([]byte(nil), o.buf[len(o.buf)-outputLogWindow:]...)
	}
	o.total += int64(len(p))
}

func (o *terminalOutput) Total() int64 {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.total
}

func (o *terminalOutput) Since(off int64) []byte {
	o.mu.Lock()
	defer o.mu.Unlock()
	start := int64(len(o.buf)) - (o.total - off)
	if start < 0 {
		start = 0
	}
	return append([]byte(nil), o.buf[start:]...)
}

// visibleText drops ANSI escapes and applies backspaces, roughly what the user saw on screen
func visibleText(p []byte) []byte {
	out := make([]byte, 0, len(p))
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch {
		case c == 0x1b && i+1 < len(p) && p[i+1] == '[':
			i += 2
			for i < len(p) && (p[i] < 0x40 || p[i] > 0x7e) {
				i++
			}
		case c == 0x1b && i+1 < len(p) && p[i+1] == ']':
			// OSC like window titles, ends with BEL or ESC \
			i += 2
			for i < len(p) && p[i] != 0x07 && !(p[i] == 0x1b && i+1 < len(p) && p[i+1] == '\\') {
				i++
			}
			if i < len(p) && p[i] == 0x1b {
				i++
			}
		case c == 0x1b:
			i++
		case c == '\b':
			if len(out) > 0 {
				_, size := utf8.DecodeLastRune(out)
				out = out[:len(out)-size]
			}
		case c == '\r' || c == 0x07:
		default:
			out = append(out, c)
		}
	}
	return out
}

type escState int

const (
	escNone escState = iota
	escStart
	escCSI
	escSS3
)

// commandRecorder rebuilds typed lines from terminal keystrokes and hands them to record once the shell echoed them back
type commandRecorder struct {
	output *terminalOutput
	record func(line string, edited bool)

	line      []byte
	edited    bool
	lineStart int64
	esc       escState
	csiParams []byte
	prevDone  chan struct{}
}

func newCommandRecorder(output *terminalOutput, record func(line string, edited bool)) *commandRecorder {
	return &commandRecorder{output: output, record: record}
}

func (c *commandRecorder) Feed(data []byte) {
	for _, b := range data {
		switch c.esc {
		case escStart:
			switch b {
			case '[':
				c.esc, c.csiParams = escCSI, c.csiParams[:0]
			case 'O':
				c.esc = escSS3
			default:
				c.esc, c.edited = escNone, true
			}
			continue
		case escCSI:
			if b >= 0x40 && b <= 0x7e {
				// bracketed paste markers are harmless, anything else is cursor movement or editing
				p := string(c.csiParams)
				if b != '~' || (p != "200" && p != "201") {
					c.edited = true
				}
				c.esc = escNone
			} else {
				c.csiParams = append(c.csiParams, b)
			}
			continue
		case escSS3:
			c.esc, c.edited = escNone, true
			continue
		}

		switch b {
		case 0x1b:
			c.esc = escStart
		case '\r', '\n':
			c.finishLine()
		case 0x7f, 0x08:
			if len(c.line) > 0 {
				_, size := utf8.DecodeLastRune(c.line)
				c.line = c.line[:len(c.line)-size]
			}
		case 0x03:
			c.resetLine()
		case 0x15:
			c.line = c.line[:0]
		case 0x17:
			trimmed := bytes.TrimRight(c.line, " ")
			if i := bytes.LastIndexByte(trimmed, ' '); i >= 0 {
				c.line = c.line[:i+1]
			} else {
				c.line = c.line[:0]
			}
		case 0x09, 0x01, 0x02, 0x05, 0x06, 0x0e, 0x10, 0x12:
			// tab, ctrl-A/B/E/F, history and reverse search keys, so the buffer may not match the real line
			c.edited = true
		default:
			if b >= 0x20 && len(c.line) < maxLoggedCommandLen {
				c.line = append(c.line, b)
			}
		}
	}
}

func (c *commandRecorder) resetLine() {
	c.line, c.edited = c.line[:0], false
	c.lineStart = c.output.Total()
}

func (c *commandRecorder) finishLine() {
	text := strings.TrimSpace(string(c.line))
	edited, start := c.edited, c.lineStart
	c.resetLine()
	if text == "" && !edited {
		return
	}
	// each check waits for the previous one so pasted lines land in the log in order
	prev, done := c.prevDone, make(chan struct{})
	c.prevDone = done
	// wait for the echo, anything typed without echo (password prompts) never gets logged
	time.AfterFunc(echoWaitDelay, func() {
		defer close(done)
		if prev != nil {
			<-prev
		}
		if text != "" && !bytes.Contains(visibleText(c.output.Since(start)), []byte(text)) {
			return
		}
		c.record(text, edited)
	})
}

func terminalCommandLogMessage(service, line string, edited bool) string {
	if line == "" {
		return "executed a command from shell history in terminal for service " + service
	}
	msg := "executed command in terminal for service " + service + ": " + line
	if edited {
		msg += " (edited with tab/arrow keys, may be incomplete)"
	}
	return msg
}
