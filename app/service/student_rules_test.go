package service

import (
	"testing"

	"siakad-mini/app/model"
)

func TestValidateStudentCreate(t *testing.T) {
	t.Run("Valid Student Request", func(t *testing.T) {
		req := model.CreateStudentRequest{
			NIM:         "434241042001",
			Nama:        "Ahmad Fauzi",
			Email:       "ahmad@student.unair.ac.id",
			Prodi:       "Sistem Informasi",
			Angkatan:    2024,
			IPKTerakhir: 3.85,
		}

		errs := ValidateStudentCreate(req)
		if len(errs) != 0 {
			t.Errorf("expected 0 validation errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("Invalid NIM Format", func(t *testing.T) {
		req := model.CreateStudentRequest{
			NIM:         "123", // invalid, should be 12 digits
			Nama:        "Budi",
			Email:       "budi@unair.ac.id",
			Prodi:       "Informatika",
			Angkatan:    2023,
			IPKTerakhir: 3.5,
		}

		errs := ValidateStudentCreate(req)
		if _, exists := errs["nim"]; !exists {
			t.Errorf("expected nim validation error, but got none")
		}
	})

	t.Run("Invalid IPK Range", func(t *testing.T) {
		req := model.CreateStudentRequest{
			NIM:         "434241042002",
			Nama:        "Citra",
			Email:       "citra@unair.ac.id",
			Prodi:       "Informatika",
			Angkatan:    2023,
			IPKTerakhir: 4.5, // invalid > 4.00
		}

		errs := ValidateStudentCreate(req)
		if _, exists := errs["ipk_terakhir"]; !exists {
			t.Errorf("expected ipk_terakhir validation error, but got none")
		}
	})
}
