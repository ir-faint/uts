package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"siakad-mini/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StudentRepository interface {
	FindByUserID(ctx context.Context, userID int) (*model.Student, error)
	FindByID(ctx context.Context, id int) (*model.Student, error)
	CreateStudentTx(ctx context.Context, tx pgx.Tx, student *model.Student) error
	UpdateStudent(ctx context.Context, id int, req *model.UpdateStudentRequest) (*model.Student, error)
	SoftDeleteStudent(ctx context.Context, id int) error
	FindAll(ctx context.Context, page, perPage int, prodi string, angkatan int, search, sort string) ([]model.StudentResponse, int, error)
	ExistsByNIM(ctx context.Context, nim string) (bool, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	GetEnrolledCoursesByStudentID(ctx context.Context, studentID int) ([]model.EnrolledCourseInfo, int, error)
}

type pgxStudentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &pgxStudentRepository{db: db}
}

func (r *pgxStudentRepository) FindByUserID(ctx context.Context, userID int) (*model.Student, error) {
	query := `SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at 
	          FROM students WHERE user_id = $1 AND deleted_at IS NULL`
	s := &model.Student{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *pgxStudentRepository) FindByID(ctx context.Context, id int) (*model.Student, error) {
	query := `SELECT s.id, s.user_id, u.email, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir, s.deleted_at 
	          FROM students s 
	          JOIN users u ON s.user_id = u.id 
	          WHERE s.id = $1 AND s.deleted_at IS NULL`
	s := &model.Student{}
	var email string
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &email, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *pgxStudentRepository) CreateStudentTx(ctx context.Context, tx pgx.Tx, student *model.Student) error {
	query := `INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return tx.QueryRow(ctx, query,
		student.UserID, student.NIM, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir,
	).Scan(&student.ID)
}

func (r *pgxStudentRepository) UpdateStudent(ctx context.Context, id int, req *model.UpdateStudentRequest) (*model.Student, error) {
	query := `UPDATE students 
	          SET nama = $1, prodi = $2, angkatan = $3, ipk_terakhir = $4 
	          WHERE id = $5 AND deleted_at IS NULL 
	          RETURNING id, user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at`
	s := &model.Student{}
	err := r.db.QueryRow(ctx, query, req.Nama, req.Prodi, req.Angkatan, req.IPKTerakhir, id).Scan(
		&s.ID, &s.UserID, &s.NIM, &s.Nama, &s.Prodi, &s.Angkatan, &s.IPKTerakhir, &s.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (r *pgxStudentRepository) SoftDeleteStudent(ctx context.Context, id int) error {
	query := `UPDATE students SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *pgxStudentRepository) FindAll(ctx context.Context, page, perPage int, prodi string, angkatan int, search, sort string) ([]model.StudentResponse, int, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	whereClauses = append(whereClauses, "s.deleted_at IS NULL")

	if prodi != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("s.prodi ILIKE $%d", argIdx))
		args = append(args, "%"+prodi+"%")
		argIdx++
	}

	if angkatan > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("s.angkatan = $%d", argIdx))
		args = append(args, angkatan)
		argIdx++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(s.nim ILIKE $%d OR s.nama ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM students s JOIN users u ON s.user_id = u.id %s", whereSQL)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	orderBy := "ORDER BY s.id ASC"
	if sort == "nama" {
		orderBy = "ORDER BY s.nama ASC"
	} else if sort == "-ipk_terakhir" {
		orderBy = "ORDER BY s.ipk_terakhir DESC"
	}

	offset := (page - 1) * perPage
	query := fmt.Sprintf(`SELECT s.id, s.user_id, u.email, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir 
		FROM students s 
		JOIN users u ON s.user_id = u.id 
		%s %s LIMIT $%d OFFSET $%d`, whereSQL, orderBy, argIdx, argIdx+1)

	args = append(args, perPage, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.StudentResponse
	for rows.Next() {
		var st model.StudentResponse
		if err := rows.Scan(&st.ID, &st.UserID, &st.Email, &st.NIM, &st.Nama, &st.Prodi, &st.Angkatan, &st.IPKTerakhir); err != nil {
			return nil, 0, err
		}
		students = append(students, st)
	}

	if students == nil {
		students = []model.StudentResponse{}
	}

	return students, total, nil
}

func (r *pgxStudentRepository) ExistsByNIM(ctx context.Context, nim string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM students WHERE nim = $1)`
	err := r.db.QueryRow(ctx, query, nim).Scan(&exists)
	return exists, err
}

func (r *pgxStudentRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

func (r *pgxStudentRepository) GetEnrolledCoursesByStudentID(ctx context.Context, studentID int) ([]model.EnrolledCourseInfo, int, error) {
	query := `SELECT e.id, e.course_id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik 
	          FROM enrollments e 
	          JOIN courses c ON e.course_id = c.id 
	          WHERE e.student_id = $1 
	          ORDER BY c.semester ASC, c.kode_mk ASC`

	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []model.EnrolledCourseInfo
	totalSKS := 0

	for rows.Next() {
		var item model.EnrolledCourseInfo
		if err := rows.Scan(&item.EnrollmentID, &item.CourseID, &item.KodeMK, &item.NamaMK, &item.SKS, &item.Semester, &item.TahunAkademik); err != nil {
			return nil, 0, err
		}
		totalSKS += item.SKS
		list = append(list, item)
	}

	if list == nil {
		list = []model.EnrolledCourseInfo{}
	}

	return list, totalSKS, nil
}
