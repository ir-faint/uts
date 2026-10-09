package service

import (
	"net/mail"
	"regexp"
	"strings"
	"time"

	"siakad-mini/app/model"
)

var nimRegex = regexp.MustCompile(`^[0-9]{12}$`)

func ValidateStudentCreate(req model.CreateStudentRequest) map[string]string {
	errs := make(map[string]string)
	currentYear := time.Now().Year()

	nim := strings.TrimSpace(req.NIM)
	if nim == "" {
		errs["nim"] = "NIM wajib diisi"
	} else if !nimRegex.MatchString(nim) {
		errs["nim"] = "NIM harus berupa 12 digit angka"
	}

	if strings.TrimSpace(req.Nama) == "" {
		errs["nama"] = "Nama wajib diisi"
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		errs["email"] = "Email wajib diisi"
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "Format email tidak valid"
	}

	if strings.TrimSpace(req.Prodi) == "" {
		errs["prodi"] = "Prodi wajib diisi"
	}

	if req.Angkatan <= 0 || req.Angkatan > currentYear {
		errs["angkatan"] = "Angkatan harus berupa tahun valid tidak melebihi tahun sekarang"
	}

	if req.IPKTerakhir < 0.00 || req.IPKTerakhir > 4.00 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}

	return errs
}

func ValidateStudentUpdate(req model.UpdateStudentRequest) map[string]string {
	errs := make(map[string]string)
	currentYear := time.Now().Year()

	if strings.TrimSpace(req.Nama) == "" {
		errs["nama"] = "Nama wajib diisi"
	}

	if strings.TrimSpace(req.Prodi) == "" {
		errs["prodi"] = "Prodi wajib diisi"
	}

	if req.Angkatan <= 0 || req.Angkatan > currentYear {
		errs["angkatan"] = "Angkatan harus berupa tahun valid tidak melebihi tahun sekarang"
	}

	if req.IPKTerakhir < 0.00 || req.IPKTerakhir > 4.00 {
		errs["ipk_terakhir"] = "IPK harus antara 0.00 dan 4.00"
	}

	return errs
}
