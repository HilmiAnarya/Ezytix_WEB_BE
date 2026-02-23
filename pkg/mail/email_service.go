package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type MailService interface {
	SendOTPEmail(toEmail string, name string, otpCode string) error
}

type mailService struct {
	dialer *gomail.Dialer
	sender string
}

func NewMailService() MailService {
	host := os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")

	port, _ := strconv.Atoi(portStr)

	dialer := gomail.NewDialer(host, port, user, pass)

	return &mailService{
		dialer: dialer,
		sender: user,
	}
}

func (s *mailService) SendOTPEmail(toEmail string, name string, otpCode string) error {
	data := struct {
		Name    string
		OTPCode string
	}{
		Name:    name,
		OTPCode: otpCode,
	}

	htmlTemplate := `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Verifikasi Akun Ezytix</title>
		<style>
			body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f9fafb; margin: 0; padding: 0; }
			.container { max-width: 600px; margin: 40px auto; background-color: #ffffff; border-radius: 12px; padding: 40px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
			.header { text-align: center; margin-bottom: 30px; }
			.logo { color: #dc2626; font-size: 32px; font-weight: bold; margin: 0; letter-spacing: -1px; }
			.logo span { color: #1f2937; }
			.content { color: #4b5563; line-height: 1.6; font-size: 16px; }
			.greeting { font-weight: bold; color: #111827; font-size: 18px; margin-bottom: 20px; }
			.otp-box { background-color: #f3f4f6; border: 2px dashed #d1d5db; border-radius: 8px; text-align: center; padding: 20px; margin: 30px 0; }
			.otp-code { font-size: 36px; font-weight: 800; color: #dc2626; letter-spacing: 8px; margin: 0; }
			.warning { font-size: 13px; color: #6b7280; text-align: center; margin-top: 10px; }
			.footer { margin-top: 40px; padding-top: 20px; border-top: 1px solid #e5e7eb; text-align: center; font-size: 13px; color: #9ca3af; }
		</style>
	</head>
	<body>
		<div class="container">
			<div class="header">
				<h1 class="logo">Ezy<span>tix</span></h1>
			</div>
			<div class="content">
				<div class="greeting">Halo, {{.Name}}! 👋</div>
				<p>Terima kasih telah mendaftar di Ezytix. Untuk menyelesaikan proses pendaftaran dan mengaktifkan akun Anda, silakan masukkan kode OTP (One-Time Password) berikut pada halaman verifikasi:</p>
				
				<div class="otp-box">
					<p class="otp-code">{{.OTPCode}}</p>
				</div>
				
				<p class="warning">⚠️ Kode ini hanya berlaku selama 5 menit. Jangan berikan kode ini kepada siapa pun, termasuk pihak Ezytix.</p>
				
				<p>Jika Anda tidak merasa melakukan pendaftaran ini, silakan abaikan email ini.</p>
			</div>
			<div class="footer">
				<p>&copy; 2025 Ezytix. All rights reserved.</p>
				<p>Surakarta, Central Java, Indonesia</p>
			</div>
		</div>
	</body>
	</html>
	`

	t, err := template.New("otp_email").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("gagal parsing template: %v", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("gagal execute template: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.sender)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Ezytix - Kode Verifikasi Akun Anda")
	m.SetBody("text/html", body.String())

	// 4. Kirim Email
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("gagal mengirim email: %v", err)
	}

	return nil
}