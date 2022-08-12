package auth

import (
	"aurora/internal/config"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore
var defaultSessionOption *sessions.Options
var users map[string]string

func Init(cfg *config.Auth) error {
	store = sessions.NewCookieStore(
		[]byte(cfg.New_authentication_key),
		[]byte(cfg.New_encryption_key),
		[]byte(cfg.Old_authentication_key),
		[]byte(cfg.Old_encryption_key),
	)
	defaultSessionOption = cfg.DefaultSessionOption
	users = cfg.Users
	return nil
}

// Authentication inspect whether session exists or not,
// If session expire, clean it
func Authentication(w http.ResponseWriter, r *http.Request) bool {
	session, err := store.Get(r, "aurora_session")
	if session.IsNew {
		http.Error(w, fmt.Sprintf("No permission %v", err), http.StatusForbidden)
		return false
	}
	// session has been damaged, clean
	if err != nil {
		session.Options = defaultSessionOption
		session.Options.MaxAge = -1 // delete imediately
		_ = session.Save(r, w)
		http.Error(w, "Clean session", http.StatusForbidden)
		return false
	}
	// session not expired
	return true
}

func DefaultStore() sessions.Store {
	return store
}