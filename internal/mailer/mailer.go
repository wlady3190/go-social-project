package mailer

import "embed"

const (
	FromName            = "WladyCorp"
	maxRetires          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

//* se debe añadir el embed de Go

//go:embed "templates"
var FS embed.FS

type Client interface {
	//*SendGrid
	// Send(templateFile, username, email string, data any, sisSandBox bool) error
	//*MailTRap
	Send(templateFile, username, email string, data any, sisSandBox bool) (int, error)
}

//* y esto va al main
