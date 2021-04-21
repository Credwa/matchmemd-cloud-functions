package matchmemdcontacts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	firebase "firebase.google.com/go"
	"github.com/sendgrid/sendgrid-go"
)

type ContactData struct {
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	PostalCode string `json:"postal_code"`
}

type ContactPutRequest struct {
	ListIds  []string      `json:"list_ids"`
	Contacts []ContactData `json:"contacts"`
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		log.Print(err)
		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintln("Request body contains badly-formed JSON")
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &malformedRequest{status: http.StatusBadRequest, msg: msg}

		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			return &malformedRequest{status: http.StatusRequestEntityTooLarge, msg: msg}

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return &malformedRequest{status: http.StatusBadRequest, msg: msg}
	}

	return nil
}

func ContactRequest(w http.ResponseWriter, r *http.Request) {
	var p ContactData
	var req ContactPutRequest
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PATCH, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "*")

	app, firebaseErr := firebase.NewApp(context.Background(), nil)
	if firebaseErr != nil {
		log.Fatalf("error initializing app: %v\n", firebaseErr)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	idToken := r.Header.Get("Authorization")
	splitToken := strings.Split(idToken, "Bearer ")
	idToken = splitToken[1]

	// Access auth service from the default app
	client, authErr := app.Auth(context.Background())
	if authErr != nil {
		log.Fatalf("error getting Auth client: %v\n", authErr)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	_, verifyErr := client.VerifyIDToken(context.Background(), idToken)
	if verifyErr != nil {
		log.Fatalf("error verifying ID token: %v\n", verifyErr)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	err := decodeJSONBody(w, r, &p)
	if err != nil {
		var mr *malformedRequest
		if errors.As(err, &mr) {
			http.Error(w, mr.msg, mr.status)
		} else {
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/marketing/contacts", "https://api.sendgrid.com")

	req.ListIds = []string{"ad881e81-938c-4721-af1b-4944cfbdee73"}
	req.Contacts = []ContactData{p}
	e, _ := json.Marshal(req)

	request.Method = "PUT"
	request.Body = e

	response, err := sendgrid.API(request)
	_ = response
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", 400)
	} else {
		http.StatusText(http.StatusOK)
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
