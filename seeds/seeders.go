package seeds

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type StudentSeedData struct {
	NIM      string
	Nama     string
	Email    string
	Prodi    string
	Angkatan int
	IPK      float64
}

type CourseSeedData struct {
	KodeMK   string
	NamaMK   string
	SKS      int
	Semester int
	Kuota    int
}

func RunSeeders(ctx context.Context, db *pgxpool.Pool) error {
	// 1. Seed Admin
	var adminCount int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&adminCount)
	if err != nil {
		return fmt.Errorf("failed to check admin count: %w", err)
	}

	if adminCount == 0 {
		hashedAdminPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		_, err = db.Exec(ctx,
			"INSERT INTO users (email, password, role) VALUES ($1, $2, $3)",
			"admin@unair.ac.id", string(hashedAdminPassword), "admin",
		)
		if err != nil {
			return fmt.Errorf("failed to seed admin user: %w", err)
		}
		fmt.Println("[Seeder] Admin user seeded successfully: admin@unair.ac.id / admin123")
	}

	// 2. Seed Mahasiswa
	var studentCount int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM students").Scan(&studentCount)
	if err != nil {
		return fmt.Errorf("failed to check student count: %w", err)
	}

	if studentCount < 20 {
		studentsData := []StudentSeedData{
			{"434241042001", "Ahmad Fauzi", "mhs1@student.unair.ac.id", "Sistem Informasi", 2023, 3.85},
			{"434241042002", "Budi Santoso", "mhs2@student.unair.ac.id", "Teknik Informatika", 2023, 3.60},
			{"434241042003", "Citra Dewi", "mhs3@student.unair.ac.id", "Sistem Informasi", 2024, 2.95},
			{"434241042004", "Dimas Prasetyo", "mhs4@student.unair.ac.id", "Teknik Informatika", 2022, 2.40},
			{"434241042005", "Eka Putri", "mhs5@student.unair.ac.id", "Data Science", 2024, 3.75},
			{"434241042006", "Fajar Nugraha", "mhs6@student.unair.ac.id", "Sistem Informasi", 2023, 3.10},
			{"434241042007", "Gita Gutawa", "mhs7@student.unair.ac.id", "Data Science", 2024, 2.80},
			{"434241042008", "Hadi Wijaya", "mhs8@student.unair.ac.id", "Teknik Informatika", 2023, 2.10},
			{"434241042009", "Indah Permata", "mhs9@student.unair.ac.id", "Sistem Informasi", 2025, 3.90},
			{"434241042010", "Joko Widodo", "mhs10@student.unair.ac.id", "Teknik Informatika", 2022, 3.25},
			{"434241042011", "Kartika Sari", "mhs11@student.unair.ac.id", "Data Science", 2024, 3.40},
			{"434241042012", "Lestari Rahayu", "mhs12@student.unair.ac.id", "Sistem Informasi", 2023, 2.70},
			{"434241042013", "Muhammad Rizky", "mhs13@student.unair.ac.id", "Teknik Informatika", 2024, 3.55},
			{"434241042014", "Nanda Pratama", "mhs14@student.unair.ac.id", "Data Science", 2023, 2.35},
			{"434241042015", "Olivia Zalianty", "mhs15@student.unair.ac.id", "Sistem Informasi", 2025, 3.80},
			{"434241042016", "Putri Utami", "mhs16@student.unair.ac.id", "Teknik Informatika", 2024, 3.15},
			{"434241042017", "Qori Sandioriva", "mhs17@student.unair.ac.id", "Data Science", 2023, 2.90},
			{"434241042018", "Rahmat Hidayat", "mhs18@student.unair.ac.id", "Sistem Informasi", 2022, 3.30},
			{"434241042019", "Siti Nurhaliza", "mhs19@student.unair.ac.id", "Teknik Informatika", 2025, 4.00},
			{"434241042020", "Taufik Hidayat", "mhs20@student.unair.ac.id", "Data Science", 2024, 2.65},
		}

		for _, s := range studentsData {
			// Check if student exists
			var exists bool
			err := db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", s.Email).Scan(&exists)
			if err != nil || exists {
				continue
			}

			tx, err := db.Begin(ctx)
			if err != nil {
				return fmt.Errorf("failed to begin tx for student seed: %w", err)
			}

			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(s.NIM), bcrypt.DefaultCost)
			if err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to hash student password for %s: %w", s.NIM, err)
			}

			var userID int
			err = tx.QueryRow(ctx,
				"INSERT INTO users (email, password, role) VALUES ($1, $2, $3) RETURNING id",
				s.Email, string(hashedPassword), "mahasiswa",
			).Scan(&userID)
			if err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to insert user for student %s: %w", s.NIM, err)
			}

			_, err = tx.Exec(ctx,
				"INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES ($1, $2, $3, $4, $5, $6)",
				userID, s.NIM, s.Nama, s.Prodi, s.Angkatan, s.IPK,
			)
			if err != nil {
				tx.Rollback(ctx)
				return fmt.Errorf("failed to insert student record %s: %w", s.NIM, err)
			}

			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("failed to commit student seed: %w", err)
			}
		}
		fmt.Println("[Seeder] 20 Mahasiswa records seeded successfully.")
	}

	// 3. Seed Courses
	var courseCount int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM courses").Scan(&courseCount)
	if err != nil {
		return fmt.Errorf("failed to check course count: %w", err)
	}

	if courseCount < 10 {
		coursesData := []CourseSeedData{
			{"IF101", "Dasar Pemrograman", 3, 1, 30},
			{"IF102", "Struktur Data & Algoritma", 4, 2, 25},
			{"IF201", "Basis Data Lanjut", 3, 3, 20},
			{"IF202", "Pemrograman Web Backend", 3, 4, 5},
			{"IF301", "Kecerdasan Buatan", 3, 5, 30},
			{"IF302", "Jaringan Komputer", 3, 5, 2},
			{"IF401", "Rekayasa Perangkat Lunak", 4, 6, 35},
			{"IF402", "Keamanan Informasi", 3, 6, 40},
			{"IF501", "Machine Learning", 3, 7, 15},
			{"IF502", "Tugas Akhir / Skripsi", 6, 8, 50},
		}

		for _, c := range coursesData {
			var exists bool
			_ = db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM courses WHERE kode_mk = $1)", c.KodeMK).Scan(&exists)
			if exists {
				continue
			}

			_, err := db.Exec(ctx,
				"INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ($1, $2, $3, $4, $5)",
				c.KodeMK, c.NamaMK, c.SKS, c.Semester, c.Kuota,
			)
			if err != nil {
				return fmt.Errorf("failed to seed course %s: %w", c.KodeMK, err)
			}
		}
		fmt.Println("[Seeder] 10 Courses seeded successfully.")
	}

	return nil
}
