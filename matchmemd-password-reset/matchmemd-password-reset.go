package matchmemdpasswordreset

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
	"firebase.google.com/go/auth"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type PasswordResetData struct {
	Email string `json:"email"`
	Host  string `json:"host"`
}

type malformedRequest struct {
	status int
	msg    string
}

func (mr *malformedRequest) Error() string {
	return mr.msg
}

func CORSEnabledFunction(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "415 - Content-Type header is not application/json unsupported", http.StatusUnsupportedMediaType)
		return nil
	}

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

func generatePasswordResetLink(p *PasswordResetData) string {
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// Access auth service from the default app
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	actionCodeSettings := &auth.ActionCodeSettings{
		URL: p.Host + "/login",
	}

	link, err := client.PasswordResetLinkWithSettings(context.Background(), p.Email, actionCodeSettings)
	if err != nil {
		log.Fatalf("error generating email link: %v\n", err)
	}

	return link
}

func dynamicTemplateEmail(pData *PasswordResetData) []byte {
	m := mail.NewV3Mail()
	link := generatePasswordResetLink(pData)

	const noReplyEmailFrom = "no-reply@matchmemd.com"

	e := mail.NewEmail(noReplyEmailFrom, noReplyEmailFrom)
	m.SetFrom(e)

	m.SetTemplateID("d-af21306d33bd4af58ab3bb3ff7536902")
	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", pData.Email),
	}

	p.AddTos(tos...)

	p.SetDynamicTemplateData("email", pData.Email)
	p.SetDynamicTemplateData("passwordResetURL", link)
	m.AddPersonalizations(p)
	return mail.GetRequestBody(m)
}

func PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	var p PasswordResetData
	CORSEnabledFunction(w, r)

	// if r.Method != http.MethodPost {
	// 	http.Error(w, "405 - Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }

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

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = dynamicTemplateEmail(&p)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong", 400)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
