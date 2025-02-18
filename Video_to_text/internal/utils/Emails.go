package utils

import (
	"fmt"
	"net/smtp"
)

func SendVerificationEmail(email, token string) error {
	// Mailtrap credentials
	from := "from@example.com"
	smtpHost := "sandbox.smtp.mailtrap.io"
	smtpPort := "2525"
	username := "ed09ce7bda3f76" // Your Mailtrap username
	password := "91e603b86da032" // Your Mailtrap password

	// Create HTML content
	htmlContent := fmt.Sprintf(`
        <!doctype html>
        <html>
            <head>
                <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
            </head>
            <body>
                <div style="display: block; margin: auto; max-width: 600px;">
                    <h1>Verify your email address</h1>
                    <p>Click the link below to verify your email:</p>
                    <a href="http://localhost:8090/auth/verify?token=%s">Verify Email</a>
                </div>
            </body>
        </html>
    `, token)

	// Create email message
	message := fmt.Sprintf(`From: %s
To: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8
Subject: Email Verification

%s`, from, email, htmlContent)

	// Setup authentication
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Send email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{email},
		[]byte(message),
	)

	return err
}

func SendVerificationPassword(email, token string) error {
	// Mailtrap credentials
	from := "from@example.com"
	smtpHost := "sandbox.smtp.mailtrap.io"
	smtpPort := "2525"
	username := "ed09ce7bda3f76" // Your Mailtrap username
	password := "91e603b86da032" // Your Mailtrap password

	// Create HTML content
	htmlContent := fmt.Sprintf(`
        <!doctype html>
        <html>
            <head>
                <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
            </head>
            <body>
                <div style="display: block; margin: auto; max-width: 600px;">
                    <h1>Verify your password address</h1>
                    <p>Click the link below to verify your email:</p>
                    <a href="http://localhost:8090/auth/ResetPassword?token=%s">Verify Email</a>
                </div>
            </body>
        </html>
    `, token)

	// Create email message
	message := fmt.Sprintf(`From: %s
To: %s
MIME-Version: 1.0
Content-Type: text/html; charset=UTF-8
Subject: Email Verification

%s`, from, email, htmlContent)

	// Setup authentication
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Send email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{email},
		[]byte(message),
	)

	return err
}
