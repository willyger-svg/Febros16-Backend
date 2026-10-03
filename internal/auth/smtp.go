package auth

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail, otp string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	if host == "" || username == "" || password == "" {
		return fmt.Errorf("SMTP credentials zimekosekana")
	}

	auth := smtp.PlainAuth("", username, password, host)

	subject := "Subject: Namba Yako ya Uthibitisho (OTP) - FEBROS16\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f4f4f5; padding: 20px;">
			<div style="max-w-md; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
				<h2 style="color: #2563eb; text-align: center;">Thibitisha Akaunti Yako</h2>
				<p style="color: #4b5563; font-size: 16px;">Habari,</p>
				<p style="color: #4b5563; font-size: 16px;">Tafadhali tumia namba hii ya uthibitisho kukamilisha usajili wako. Namba hii itaisha muda wake ndani ya dakika 15.</p>
				<div style="text-align: center; margin: 30px 0;">
					<span style="background-color: #f3f4f6; color: #1f2937; padding: 16px 32px; border-radius: 8px; font-weight: 900; font-size: 28px; letter-spacing: 4px; display: inline-block; border: 2px dashed #cbd5e1;">%s</span>
				</div>
				<p style="color: #6b7280; font-size: 14px;">Kama hukufanya jaribio hili, tafadhali puuza ujumbe huu.</p>
				<hr style="border: 0; border-top: 1px solid #e5e7eb; margin: 20px 0;" />
				<p style="color: #9ca3af; font-size: 12px; text-align: center;">© 2026 FEBROS16 Portal.</p>
			</div>
		</body>
		</html>
	`, otp)

	msg := []byte(subject + mime + body)

	addr := fmt.Sprintf("%s:%s", host, port)
	return smtp.SendMail(addr, auth, username, []string{toEmail}, msg)
}
