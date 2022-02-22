package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type AuthRequest struct {
	Name     string `json:"name" validate:"required,gt=0,lt=15"`
	Password string `json:"password" validate:"required,gt=0,lt=30"`
}

type AuthResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
}

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Must use POST", http.StatusBadRequest)
		return
	}

	session, err := store.Get(r, "aurora_session")

	// Session not expired
	{
		// if err == nil && session.IsNew == false {
		// 	w.WriteHeader(http.StatusOK)
		// 	responseOBJ := &AuthResponse{
		// 		Message: "login repeat",
		// 	}
		// 	data, err := json.Marshal(responseOBJ)
		// 	if err != nil {
		// 		http.Error(w, fmt.Sprintf("Failed to json.Marshal responseOBJ: %v", err), http.StatusBadGateway)
		// 		return
		// 	}
		// 	w.WriteHeader(http.StatusOK)
		// 	w.Write(data)
		// 	return
		// }
	}

	// Obtain failed or Create a new session
	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
	strByte, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read req.Body failed", http.StatusBadRequest)
		return
	}

	requestOBJ := &AuthRequest{}
	decoder := json.NewDecoder(bytes.NewReader(strByte))
	decoder.UseNumber()

	if err := decoder.Decode(requestOBJ); err != nil {
		http.Error(w, fmt.Sprintf("Unexpected request Unmarshal format: %v", err), http.StatusBadRequest)
		return
	}

	if err := Validate(requestOBJ); err != nil {
		http.Error(w, fmt.Sprintf("Failed to validate format: %v", err), http.StatusForbidden)
		return
	}

	// Login successful
	session.Options = defaultSessionOption
	uuid := uuid.New().String()
	session.Values["UUID"] = uuid
	session.Values["User"] = requestOBJ.Name

	if err := session.Save(r, w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseOBJ := &AuthResponse{
		Message: "login successful",
		Name:    requestOBJ.Name,
		UUID:    uuid,
	}

	data, err := json.Marshal(responseOBJ)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to json.Marshal responseOBJ: %v", err), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Validate parameters and name & password whether pass or not
func Validate(requestOBJ *AuthRequest) error {
	if err := validate.Struct(requestOBJ); err != nil {
		return err
	}
	if password, ok := users[requestOBJ.Name]; ok && requestOBJ.Password == password {
		return nil
	}
	return errors.New("Name and Password not match")
}
