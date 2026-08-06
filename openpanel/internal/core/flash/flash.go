// Package flash implements one-shot flash messages: a message stashed in
// the session and popped on the next render, used for things like "IP
// address mismatch, please login again."
package flash

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
)

// Message is a single flashed (category, text) pair.
type Message struct {
	Category string
	Text     string
}

func init() {
	gob.Register(Message{})
}

// Add stashes a flash message in the session. The session must be saved by the
// caller (store.Save(r, w, sess)) for this to persist.
func Add(sess *sessions.Session, category, message string) {
	sess.AddFlash(Message{Category: category, Text: message})
}

// Pop reads and clears all flashed messages. It also saves the session so
// the messages don't reappear on the next request.
func Pop(store sessions.Store, w http.ResponseWriter, r *http.Request, sess *sessions.Session) []Message {
	raw := sess.Flashes()
	if len(raw) > 0 {
		_ = store.Save(r, w, sess)
	}

	messages := make([]Message, 0, len(raw))
	for _, v := range raw {
		if m, ok := v.(Message); ok {
			messages = append(messages, m)
		}
	}
	return messages
}
