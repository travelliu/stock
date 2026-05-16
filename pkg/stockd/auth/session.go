package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
)

const SessionName = "stockd_session"

// NewSessionStore returns a signed cookie store with sensible defaults.
// secure should match whether the server is running TLS; setting it to true
// over plain HTTP causes browsers to silently drop the cookie after login.
// secret must be >= 32 bytes (enforced by config validation).
func NewSessionStore(secret []byte, secure bool) sessions.Store {
	store := cookie.NewStore(secret)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   secure,
		SameSite: 2, // http.SameSiteLaxMode
	})
	return store
}
