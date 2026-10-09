package model

import "time"

type Student struct {
	ID          int        `json:"id"`
	UserID      int        `json:"user_id"`
	NIM         string     `json:"nim"`
	Nama        string     `json:"nama"`
	Prodi       string     `json:"prodi"`
	Angkatan    int        `json:"angkatan"`
	IPKTerakhir float64    `json:"ipk_terakhir"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type StudentDetail struct {
	NIM      string  `json:"nim"`
	Nama     string  `json:"nama"`
	Prodi    string  `json:"prodi"`
	Angkatan int     `json:"angkatan"`
	IPK      *float64 `json:"ipk_terakhir,omitempty"`
}

type StudentResponse struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Email       string  `json:"email,omitempty"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type StudentDetailResponse struct {
	ID              int                   `json:"id"`
	UserID          int                   `json:"user_id"`
	Email           string                `json:"email,omitempty"`
	NIM             string                `json:"nim"`
	Nama            string                `json:"nama"`
	Prodi           string                `json:"prodi"`
	Angkatan        int                   `json:"angkatan"`
	IPKTerakhir     float64               `json:"ipk_terakhir"`
	BatasSKS        int                   `json:"batas_sks"`
	TotalSKSDiambil int                   `json:"total_sks_diambil"`
	EnrolledCourses []EnrolledCourseInfo  `json:"enrolled_courses"`
}

type CreateStudentRequest struct {
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Email       string  `json:"email"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type UpdateStudentRequest struct {
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}


