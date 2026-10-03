package auth

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")

	if smtpHost == "" || username == "" || password == "" {
		return fmt.Errorf("SMTP credentials zimekosekana")
	}

	fromHeader := fmt.Sprintf("FEBROS16 <%s>", username)
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\n", fromHeader, toEmail)
	subject := "Subject: Namba Yako ya Uthibitisho (OTP) - FEBROS16\r\n"
	mime := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
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

	msg := []byte(headers + subject + mime + body)

	// TLS Config ya moja kwa moja
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         smtpHost,
	}

	// Dialing connection ya TLS kupitia port 465
	conn, err := tls.Dial("tcp", smtpHost+":"+smtpPort, tlsConfig)
	if err != nil {
		log.Printf("🔥 TLS Dial failed: %v", err)
		return fmt.Errorf("TLS Dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		log.Printf("🔥 SMTP client creation failed: %v", err)
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Quit()

	// CRITICAL FIX: Lazimisha kujitambulisha na Host badala ya localhost
	if err = client.Hello(smtpHost); err != nil {
		return fmt.Errorf("HELO handshake failed: %w", err)
	}

	auth := smtp.PlainAuth("", username, password, smtpHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP Authentication failed: %w", err)
	}

	if err = client.Mail(username); err != nil {
		return fmt.Errorf("MAIL FROM command failed: %w", err)
	}
	
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("RCPT TO command failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA command initialization failed: %w", err)
	}
	
	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("Writing data body failed: %w", err)
	}
	
	err = w.Close()
	if err != nil {
		return fmt.Errorf("Closing data writer failed: %w", err)
	}

	log.Println("✅ Email imetumwa kikamilifu kwa:", toEmail)
	return nil
}
