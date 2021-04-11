package matchmemdpasswordreset

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// "context"
	// firebase "firebase.google.com/go"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// func main() {

// 	app, err := firebase.NewApp(context.Background(), nil)
// 	if err != nil {
// 		log.Fatalf("error initializing app: %v\n", err)
// 	}
// 	log.Print(app)
// }

func PasswordResetRequest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello Beautiful World!\n")
	from := mail.NewEmail("Example User", "admin@matchmemd.com")
	subject := "Sending with SendGrid is Fun"
	to := mail.NewEmail("Example User", "craigroe7@gmail.com")
	plainTextContent := "and easy to do anywhere, even with Go"
	htmlContent := "<strong>and easy to do anywhere, even with Go</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	log.Println(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)

		log.Println(response.StatusCode)
		log.Println(response.Body)
		log.Println(response.Headers)
	}
}
