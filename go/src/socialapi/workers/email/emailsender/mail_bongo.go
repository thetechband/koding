// Package sender provides an API for mail sending operations
package emailsender

// Mail struct hold the required parameters for sending an email
type Mail struct {
	To      string
	ToName  string
	Subject string
	Text    string
	// TODO maybe we can remove this HTML field
	HTML       string
	From       string
	Bcc        string
	FromName   string
	ReplyTo    string
	Properties *Properties
}

type Properties struct {
	Username     string
	Substituions map[string]string
}

func NewMail(to, from, subject, username string) *Mail {
	substituions := map[string]string{}
	properties := &Properties{Username: username, Substituions: substituions}

	return &Mail{
		To: to, From: from, Subject: subject,
		Properties: properties,
	}
}

func (m Mail) GetId() int64 {
	return 0
}

func (m Mail) BongoName() string {
	return "api.mail"
}
