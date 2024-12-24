package mailer

import (
	"bytes"
	"fmt"
	"html/template"

	"log"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailer struct {
	fromEmail string
	apikey    string
	client    *sendgrid.Client
}

func NewSendgrid(apikey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apikey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apikey:    apikey,
		client:    client,
	}
}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandBox bool) error {
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	//* template parsing and building

	tmpl, err := template.ParseFS(FS, "templates/"+UserWelcomeTemplate)

	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)

	if err != nil {
		return err
	}
	body := new(bytes.Buffer)

	err = tmpl.ExecuteTemplate(body, "body", data)

	if err != nil {
		return err
	}

	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandBox,
		},
	})

	for i := 0; i < maxRetires; i++ {
		response, err := m.client.Send(message)
		if err != nil {
			log.Printf("failed to send email to %v, attempt %d of %d", email, i+1, maxRetires)
			log.Printf("Error: %v", err.Error())
			//exponential backoff
			time.Sleep(time.Second * time.Duration(i+1))
			continue //para que salte a la siguiente iteracion si hay error

		}

		log.Printf("email sent to %s with status code %v", email, response.StatusCode)
		return nil
	}
	return fmt.Errorf("failed after %d attempts", maxRetires)

} //* Y esto va al main -> main()
