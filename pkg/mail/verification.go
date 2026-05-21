package mail

import (
	"fmt"
	"os"
	"strconv"
)

func SendVerificationEmail(to, email, otpCode, verificationLink string) error {
	expiryMinutes, err := strconv.Atoi(os.Getenv("EXPIRED_OTP"))
	if err != nil || expiryMinutes <= 0 {
		expiryMinutes = 5
	}

	htmlBody, err := RenderVerificationEmail(VerificationEmailData{
		Name:             email,
		Code:             otpCode,
		VerificationLink: verificationLink,
		ExpiryMinutes:    expiryMinutes,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	subject := "Verifikasi Akun - GreenTrust"
	return SendEmail(to, subject, htmlBody)

}
