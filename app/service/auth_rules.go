package service

import (
	"net/mail"
	"strings"

	"siakad-mini/app/model"
)

func ValidateLogin(req model.LoginRequest) map[string]string {
	errs := make(map[string]string)

	email := strings.TrimSpace(req.Email)
	if email == "" {
		errs["email"] = "Email wajib diisi"
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "Format email tidak valid"
	}

	if req.Password == "" {
		errs["password"] = "Password wajib diisi"
	} else if len(req.Password) < 8 {
		errs["password"] = "Password minimal 8 karakter"
	}

	return errs
}
