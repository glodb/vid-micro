package emails

import (
	"sync"

	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
	"gopkg.in/gomail.v2"
)

var (
	once     sync.Once
	instance *emailSender
)

type emailSender struct {
	sender *gomail.Dialer
}

func GetInstance() *emailSender {
	once.Do(func() {
		// Initialize the email sender with your email configuration.

		instance = &emailSender{

			sender: gomail.NewDialer(
				configmanager.GetInstance().EmailConfig.SMTPServer,
				configmanager.GetInstance().EmailConfig.Port,
				configmanager.GetInstance().EmailConfig.Username,
				configmanager.GetInstance().EmailConfig.Password,
			),
		}
	})
	return instance
}

func (e *emailSender) SendVerificationEmail(to, subject, body string) error {
	// Create a new message.
	message := gomail.NewMessage()
	message.SetHeader("From", e.sender.Username)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/plain", body)

	// Send the email.
	if err := e.sender.DialAndSend(message); err != nil {
		return err
	}

	return nil
}
