package helper

import (
	"context"
	"strconv"
	"strings"
	"time"

	"siakad-mini/app/model"

	"github.com/gofiber/fiber/v2"
)

// RequestContext gives context with timeout for database operations (5s default)
func RequestContext(c *fiber.Ctx) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.UserContext(), 5*time.Second)
}

// ParamID parses parameter :id from route and validates positive integer
func ParamID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}

var allowedSort = map[string]bool{
	"id":            true,
	"nama":          true,
	"ipk_terakhir":  true,
	"-ipk_terakhir": true,
}

// ParseListQuery reads query string parameters and assigns safe defaults
func ParseListQuery(c *fiber.Ctx) model.ListQuery {
	angkatan, _ := strconv.Atoi(c.Query("angkatan", "0"))

	q := model.ListQuery{
		Page:     c.QueryInt("page", 1),
		Limit:    c.QueryInt("per_page", c.QueryInt("limit", 10)),
		Search:   strings.TrimSpace(c.Query("search")),
		Sort:     c.Query("sort", "id"),
		Prodi:    strings.TrimSpace(c.Query("prodi")),
		Angkatan: angkatan,
	}

	if q.Page < 1 {
		q.Page = 1
	}
	if q.Limit < 1 {
		q.Limit = 10
	}
	if q.Limit > 50 {
		q.Limit = 50
	}

	if !allowedSort[q.Sort] {
		q.Sort = "id"
	}

	return q
}
