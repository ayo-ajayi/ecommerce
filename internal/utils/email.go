package utils

import (
	"bytes"
	"html/template"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailManager struct {
	SenderEmail string
	SenderName  string
	ApiKey      string
}

func NewEmailManager(senderEmail, senderName, apiKey string) *EmailManager {
	return &EmailManager{
		SenderEmail: senderEmail,
		SenderName:  senderName,
		ApiKey:      apiKey,
	}
}
func (eu *EmailManager) sendEmail(subject, email, firstname, otp, title, h1, p string) error {
	from := mail.NewEmail(eu.SenderName, eu.SenderEmail)
	to := mail.NewEmail(firstname, email)
	htmlContent, err := eu.emailHTML(otp, firstname, title, h1, p)
	if err != nil {
		return err
	}
	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(eu.ApiKey)
	_, err = client.Send(message)
	return err
}

func (eu *EmailManager) emailHTML(otp, firstname, title, h1, p string) (string, error) {
	tmpl := template.Must(template.New("email").Parse(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>{{.Title}}</title>
			<style>
				body {
					font-family: Arial, sans-serif;
					background-color: #f5f5f5;
					margin: 0;
					padding: 0;
					display: flex;
					justify-content: center;
					align-items: center;
					min-height: 100vh;
				}
				.container {
					background-color: #ffffff;
					border: 1px solid #e0e0e0;
					border-radius: 8px;
					box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
					padding: 20px;
					text-align: center;
					max-width: 400px;
					margin: 0 auto;
				}
				h1 {
					color: #333;
				}
				p {
					color: #666;
				}
				.otp {
					font-size: 24px;
					color: #007bff;
				}
				.footer {
					margin-top: 20px;
					color: #999;
				}
				.firstname {
					font-weight: bold;
					color: #808080;
				}
			</style>
		</head>
		<body>
			<div class="container">
				<h1>{{.H1}}</h1>
				<p>Dear <span class="firstname">{{.Firstname}}</span>,</p>
				<p>{{.P}}</p>
				<p class="otp">Your OTP: {{.Otp}}</p>
				<p class="footer">This email was sent by {{.Sendername}}</p>
			</div>
		</body>
	</html>
	`))

	var buf bytes.Buffer
	data := struct {
		Title      string
		H1         string
		Firstname  string
		P          string
		Otp        string
		Sendername string
	}{
		Title:      title,
		H1:         h1,
		Firstname:  firstname,
		P:          p,
		Otp:        otp,
		Sendername: eu.SenderName,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (eu *EmailManager) SendAccountVerificationEmail(email, firstname, otp string) error {
	subject := "Verify your " + eu.SenderName + " account"
	title := "Email Verification"
	h1 := "Email Verification"
	p := "Thank you for signing up! Please verify your email to activate your account."
	return eu.sendEmail(subject, email, firstname, otp, title, h1, p)
}

func (eu *EmailManager) SendResetPasswordEmail(email, firstname, otp string) error {
	subject := "Reset your " + eu.SenderName + " account password"
	title := "Password Reset"
	h1 := "Password Reset"
	p := "Hiii! Please enter the OTP to reset your password."
	return eu.sendEmail(subject, email, firstname, otp, title, h1, p)
}
