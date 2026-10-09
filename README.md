# SIAKAD Mini - Production-Ready Academic RESTful API Backend

A production-ready RESTful API backend for the **SIAKAD Mini** academic service built with **Go (Go Fiber v2)**, **PostgreSQL (`pgx/v5`)**, and strict **Clean Architecture** as specified in the course modules.

---

## 🏛️ Architecture & Project Structure (Lecturer Module Compliant)

This project strictly follows the **4-Layer Clean Architecture** & **Dependency Rule** defined in Modul 4 & 5:

```text
uts/
├── app/
│   ├── model/                  # Layer 1: Entities, DTOs (NO SQL, NO Fiber, NO internal imports)
│   │   ├── course.go
│   │   ├── enrollment.go
│   │   ├── response.go
│   │   ├── student.go
│   │   └── user.go
│   ├── repository/             # Layer 3: Database access (PostgreSQL pgx/v5, NO Fiber imports)
│   │   ├── course_repository.go
│   │   ├── enrollment_repository.go
│   │   ├── student_repository.go
│   │   └── user_repository.go
│   └── service/                # Layer 2 & 3: Use cases, HTTP handlers, & pure business rules
│       ├── auth_rules.go       # Pure validation & rules without fiber.Ctx (unit-testable)
│       ├── auth_service.go
│       ├── course_service.go
│       ├── enrollment_rules.go # Pure enrollment rules
│       ├── enrollment_service.go
│       ├── student_rules.go    # Pure student validation rules
│       ├── student_rules_test.go # Unit tests (`go test ./app/service/...`)
│       └── student_service.go
├── config/                     # Layer 4: Configuration & Application Assembly
│   ├── app.go                  # Fiber instance, error handler, route registration
│   ├── config.go               # Config struct
│   ├── env.go                  # Environment variable helpers
│   └── logger.go               # Structured slog JSON logger + lumberjack log rotator
├── database/                   # Layer 4: Database connection pool & auto-migrations
│   └── database.go
├── helper/                     # Layer 3: Cross-package presenter & utility helpers
│   ├── context.go              # LocalsAuthUser context helper
│   ├── helper.go               # CalculateBatasSKS pure function
│   ├── jwt.go                  # JWT token generation & parsing
│   ├── request.go              # RequestContext (5s timeout) & ParamID helpers
│   ├── response.go             # Presenters: Success, SuccessList, Created, Fail, FailValidation
│   └── security.go             # Password hashing (bcrypt) & login rate limiter
├── logs/                       # Log file output folder (ignored in git)
│   └── app.log
├── middleware/                 # Layer 4: Fiber middlewares (RequireAuth, RequireRole, RequireJSON, RequestLogger)
│   ├── auth.go
│   └── middleware.go
├── migrations/                 # Embedded database schema migration SQL files
│   ├── 000001_init_schema.up.sql
│   └── migrations.go
├── route/                      # Layer 4: Route registrations mapping URLs to services (NO business logic)
│   └── route.go
├── seeds/                      # Initial database seeders (1 Admin, 20 Mahasiswa, 10 Courses)
│   └── seeders.go
├── .env.example
├── .gitignore                  # Git ignore rules (includes .env & logs/)
├── go.mod
├── go.sum
├── main.go                     # Entrypoint: Config loading, app assembly, & Graceful Shutdown
└── README.md
```

---

## 🔒 Layer Dependency Rule Check (Zero Import Cycles)

- `app/model` -> imports **nothing** from own project.
- `app/repository` -> imports **`app/model` only** (100% free of Fiber imports!).
- `app/service` -> imports `app/model`, `app/repository`, `helper`.
- `helper` -> imports `app/model`.
- `middleware` -> imports `helper`, `app/model`.
- `route` -> imports `app/service`, `middleware`, `helper`.
- `config` -> imports `app/service`, `route`, `middleware`, `helper`.
- `database` -> imports `config`.
- `main.go` -> imports `config`, `database`, `app/repository`, `app/service`.

---

## 🧪 Running Tests & Build Verification

Run unit tests for pure business rules:
```bash
go test ./app/service/... -v
```

Check code quality and layer isolation:
```bash
go vet ./...
```

Run application:
```bash
go run main.go
```
