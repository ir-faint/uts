package repository

import (
	"context"
	"errors"

	"siakad-mini/app/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EnrollmentRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	CountActiveEnrollmentsTx(ctx context.Context, tx pgx.Tx, courseID int) (int, error)
	IsAlreadyEnrolledTx(ctx context.Context, tx pgx.Tx, studentID, courseID int, tahunAkademik string) (bool, error)
	GetEnrolledSKSInTahunAkademikTx(ctx context.Context, tx pgx.Tx, studentID int, tahunAkademik string) (int, error)
	CreateEnrollmentTx(ctx context.Context, tx pgx.Tx, enrollment *model.Enrollment) error
	FindByID(ctx context.Context, id int) (*model.Enrollment, error)
	DeleteEnrollment(ctx context.Context, id int) error
}

type pgxEnrollmentRepository struct {
	db *pgxpool.Pool
}

func NewEnrollmentRepository(db *pgxpool.Pool) EnrollmentRepository {
	return &pgxEnrollmentRepository{db: db}
}

func (r *pgxEnrollmentRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.db.Begin(ctx)
}

func (r *pgxEnrollmentRepository) CountActiveEnrollmentsTx(ctx context.Context, tx pgx.Tx, courseID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`
	err := tx.QueryRow(ctx, query, courseID).Scan(&count)
	return count, err
}

func (r *pgxEnrollmentRepository) IsAlreadyEnrolledTx(ctx context.Context, tx pgx.Tx, studentID, courseID int, tahunAkademik string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM enrollments WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3)`
	err := tx.QueryRow(ctx, query, studentID, courseID, tahunAkademik).Scan(&exists)
	return exists, err
}

func (r *pgxEnrollmentRepository) GetEnrolledSKSInTahunAkademikTx(ctx context.Context, tx pgx.Tx, studentID int, tahunAkademik string) (int, error) {
	var totalSKS int
	query := `
		SELECT COALESCE(SUM(c.sks), 0) 
		FROM enrollments e 
		JOIN courses c ON e.course_id = c.id 
		WHERE e.student_id = $1 AND e.tahun_akademik = $2
	`
	err := tx.QueryRow(ctx, query, studentID, tahunAkademik).Scan(&totalSKS)
	return totalSKS, err
}

func (r *pgxEnrollmentRepository) CreateEnrollmentTx(ctx context.Context, tx pgx.Tx, enrollment *model.Enrollment) error {
	query := `INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES ($1, $2, $3) RETURNING id, created_at`
	return tx.QueryRow(ctx, query, enrollment.StudentID, enrollment.CourseID, enrollment.TahunAkademik).Scan(&enrollment.ID, &enrollment.CreatedAt)
}

func (r *pgxEnrollmentRepository) FindByID(ctx context.Context, id int) (*model.Enrollment, error) {
	query := `SELECT id, student_id, course_id, tahun_akademik, created_at FROM enrollments WHERE id = $1`
	e := &model.Enrollment{}
	err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.StudentID, &e.CourseID, &e.TahunAkademik, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return e, nil
}

func (r *pgxEnrollmentRepository) DeleteEnrollment(ctx context.Context, id int) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
