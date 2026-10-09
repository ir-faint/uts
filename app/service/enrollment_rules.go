package service

import (
	"strings"

	"siakad-mini/app/model"
)

func ValidateEnrollment(req model.CreateEnrollmentRequest) map[string]string {
	errs := make(map[string]string)

	if req.CourseID <= 0 {
		errs["course_id"] = "course_id valid wajib diisi"
	}

	if strings.TrimSpace(req.TahunAkademik) == "" {
		errs["tahun_akademik"] = "tahun_akademik wajib diisi"
	}

	return errs
}
