package request

import (
	"aurora/internal/auth"
	"errors"
	"net/http"

	validator "github.com/go-playground/validator/v10"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Verification parameters
func (f *CenterRequest) Validate() (err error) {
	if err = validate.Struct(f); err != nil {
		return err
	}
	return nil
}

// Inject will add some session attribute in f
func (f *CenterRequest) Inject(r *http.Request) (err error) {
	// Inject session`s attribute in f
	store := auth.DefaultStore()
	if store == nil {
		return errors.New("Server Session Store not Init")
	}
	session, _ := store.Get(r, "aurora_session")
	if session.IsNew {
		return errors.New("Read session from request exceptly")
	}
	f.User = session.Values["User"].(string)
	f.UUID = session.Values["UUID"].(string)
	return nil
}
