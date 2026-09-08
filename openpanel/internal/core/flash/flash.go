// Package flash implements one-shot flash messages: stashed in the session, popped on the next render.
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

// Add stashes a flash message in the session; the caller still needs to save the session for it to persist
func Add(sess *sessions.Session, category, message string) {
	sess.AddFlash(Message{Category: category, Text: message})
}

// Pop reads and clears all flashed messages, saving the session so they don't reappear next request
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
