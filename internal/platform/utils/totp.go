package utils

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func GenerateTOTPSecret(userEmail string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      Issuer,
		AccountName: userEmail,
	})
}

func IsTOTPValid(presentedTOTP string, secret string) bool {
	return totp.Validate(presentedTOTP, secret)
}
