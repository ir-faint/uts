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

type CourseRepository interface {
	FindAll(ctx context.Context, semester int, search string, availableOnly bool) ([]model.CourseResponse, error)
	FindByIDForUpdateTx(ctx context.Context, tx pgx.Tx, courseID int) (*model.Course, error)
	FindByID(ctx context.Context, courseID int) (*model.Course, error)
}

type pgxCourseRepository struct {
	db *pgxpool.Pool
}

func NewCourseRepository(db *pgxpool.Pool) CourseRepository {
	return &pgxCourseRepository{db: db}
}

func (r *pgxCourseRepository) FindAll(ctx context.Context, semester int, search string, availableOnly bool) ([]model.CourseResponse, error) {
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	if semester > 0 {
		whereClauses = append(whereClauses, fmt.Sprintf("c.semester = $%d", argIdx))
		args = append(args, semester)
		argIdx++
	}

	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(c.kode_mk ILIKE $%d OR c.nama_mk ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	havingSQL := ""
	if availableOnly {
		havingSQL = "HAVING (c.kuota - COUNT(e.id)) > 0"
	}

	query := fmt.Sprintf(`
		SELECT 
			c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota,
			COUNT(e.id)::int AS terisi,
			(c.kuota - COUNT(e.id))::int AS sisa_kuota
		FROM courses c
		LEFT JOIN enrollments e ON c.id = e.course_id
		%s
		GROUP BY c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota
		%s
		ORDER BY c.semester ASC, c.kode_mk ASC
	`, whereSQL, havingSQL)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []model.CourseResponse
	for rows.Next() {
		var c model.CourseResponse
		if err := rows.Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota, &c.Terisi, &c.SisaKuota); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}

	if courses == nil {
		courses = []model.CourseResponse{}
	}

	return courses, nil
}

func (r *pgxCourseRepository) FindByIDForUpdateTx(ctx context.Context, tx pgx.Tx, courseID int) (*model.Course, error) {
	query := `SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1 FOR UPDATE`
	c := &model.Course{}
	err := tx.QueryRow(ctx, query, courseID).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}

func (r *pgxCourseRepository) FindByID(ctx context.Context, courseID int) (*model.Course, error) {
	query := `SELECT id, kode_mk, nama_mk, sks, semester, kuota FROM courses WHERE id = $1`
	c := &model.Course{}
	err := r.db.QueryRow(ctx, query, courseID).Scan(&c.ID, &c.KodeMK, &c.NamaMK, &c.SKS, &c.Semester, &c.Kuota)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return c, nil
}
