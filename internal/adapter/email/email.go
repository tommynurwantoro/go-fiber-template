package email

import (
	"app/config"
	"fmt"

	"gopkg.in/gomail.v2"
)

type EmailAdapter interface {
	SendEmail(to, subject, body string) error
	SendResetPasswordEmail(to, token string) error
	SendVerificationEmail(to, token string) error
}

type EmailAdapterImpl struct {
	Conf   *config.Config `inject:"config"`
	Dialer *gomail.Dialer
}

func (a *EmailAdapterImpl) Startup() error {
	a.Dialer = gomail.NewDialer(
		a.Conf.SMTP.Host,
		a.Conf.SMTP.Port,
		a.Conf.SMTP.Username,
		a.Conf.SMTP.Password,
	)
	return nil
}

func (a *EmailAdapterImpl) Shutdown() error {
	return nil
}

func (a *EmailAdapterImpl) SendEmail(to, subject, body string) error {
	mailer := gomail.NewMessage()
	mailer.SetHeader("From", a.Conf.SMTP.From)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	if err := a.Dialer.DialAndSend(mailer); err != nil {
		return err
	}

	return nil
}

func (a *EmailAdapterImpl) SendResetPasswordEmail(to, token string) error {
	subject := "Reset password"

	// TODO: replace this url with the link to the reset password page of your front-end app
	resetPasswordURL := fmt.Sprintf("http://link-to-app/reset-password?token=%s", token)
	body := fmt.Sprintf(`Dear user,

To reset your password, click on this link: %s

If you did not request any password resets, then ignore this email.`, resetPasswordURL)
	return a.SendEmail(to, subject, body)
}

func (a *EmailAdapterImpl) SendVerificationEmail(to, token string) error {
	subject := "Email Verification"

	// TODO: replace this url with the link to the email verification page of your front-end app
	verificationEmailURL := fmt.Sprintf("http://link-to-app/verify-email?token=%s", token)
	body := fmt.Sprintf(`Dear user,

To verify your email, click on this link: %s

If you did not create an account, then ignore this email.`, verificationEmailURL)
	return a.SendEmail(to, subject, body)
}
