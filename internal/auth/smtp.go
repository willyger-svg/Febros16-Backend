package auth

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(toEmail, token string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	if host == "" || username == "" || password == "" {
		return fmt.Errorf("SMTP credentials zimekosekana kwenye .env")
	}

	auth := smtp.PlainAuth("", username, password, host)

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	
	// Tunatumia Backend API URL kufanya verification ili database i-update moja kwa moja
	// Au tunaweza tumia Frontend URL ambayo ita-hit backend. Kwa urahisi tutumie Backend ku-verify then i-redirect
	backendURL := "https://febros16-backend.onrender.com"
	if os.Getenv("ENV") == "development" {
		backendURL = "http://localhost:8080"
	}

	verifyLink := fmt.Sprintf("%s/api/v1/auth/verify?token=%s", backendURL, token)

	subject := "Subject: Thibitisha Barua Pepe Yako - FEBROS16\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f5; padding: 20px;">
			<div style="max-w-md; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
				<h2 style="color: #2563eb; text-align: center;">Karibu FEBROS16!</h2>
				<p style="color: #4b5563; font-size: 16px;">Habari,</p>
				<p style="color: #4b5563; font-size: 16px;">Asante kwa kujiunga na Jukwaa la Watafiti (FEBROS16). Tafadhali bofya kitufe hapa chini kuthibitisha barua pepe yako:</p>
				<div style="text-align: center; margin: 30px 0;">
					<a href="%s" style="background-color: #2563eb; color: white; padding: 12px 24px; text-decoration: none; border-radius: 6px; font-weight: bold; display: inline-block;">Thibitisha Akaunti</a>
				</div>
				<p style="color: #6b7280; font-size: 14px;">Kama hukufanya jaribio hili, tafadhali puuza ujumbe huu.</p>
				<hr style="border: 0; border-top: 1px solid #e5e7eb; margin: 20px 0;" />
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">© 2026 FEBROS16 Portal.</p>
			</div>
		</body>
		</html>
	`, verifyLink)

	msg := []byte(subject + mime + body)

	addr := fmt.Sprintf("%s:%s", host, port)
	return smtp.SendMail(addr, auth, username, []string{toEmail}, msg)
}
