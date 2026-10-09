package model

import "time"

type Enrollment struct {
	ID            int       `json:"id"`
	StudentID     int       `json:"student_id"`
	CourseID      int       `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateEnrollmentRequest struct {
	CourseID      int    `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

type EnrollmentResponse struct {
	ID            int             `json:"id"`
	StudentID     int             `json:"student_id"`
	CourseID      int             `json:"course_id"`
	TahunAkademik string          `json:"tahun_akademik"`
	CreatedAt     time.Time       `json:"created_at"`
	Course        *CourseResponse `json:"course,omitempty"`
}

type EnrolledCourseInfo struct {
	EnrollmentID  int    `json:"enrollment_id"`
	CourseID      int    `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}
