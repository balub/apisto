package api

import (
	"context"

	"github.com/balub/apisto/internal/database"
	"github.com/balub/apisto/internal/models"
	"github.com/gofiber/fiber/v2"
)

type projectHandlers struct {
	db *database.DB
}

func (h *projectHandlers) create(c *fiber.Ctx) error {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errResponse(c, 400, "invalid request body", "INVALID_INPUT")
	}
	if body.Name == "" {
		return errResponse(c, 400, "name is required", "INVALID_INPUT")
	}

	var p models.Project
	err := h.db.Pool.QueryRow(context.Background(), `
		INSERT INTO projects (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at`,
		body.Name, body.Description,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return errResponse(c, 500, "failed to create project", "INTERNAL_ERROR")
	}
	return c.Status(201).JSON(p)
}

func (h *projectHandlers) list(c *fiber.Ctx) error {
	rows, err := h.db.Pool.Query(context.Background(), `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at,
		       COUNT(d.id) AS device_count
		FROM projects p
		LEFT JOIN devices d ON d.project_id = p.id
		GROUP BY p.id ORDER BY p.created_at DESC`)
	if err != nil {
		return errResponse(c, 500, "failed to list projects", "INTERNAL_ERROR")
	}
	defer rows.Close()

	var projects []models.ProjectWithStats
	for rows.Next() {
		var p models.ProjectWithStats
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.DeviceCount); err != nil {
			return errResponse(c, 500, "scan error", "INTERNAL_ERROR")
		}
		projects = append(projects, p)
	}
	if projects == nil {
		projects = []models.ProjectWithStats{}
	}
	return c.JSON(projects)
}

func (h *projectHandlers) get(c *fiber.Ctx) error {
	id := c.Params("id")
	var p models.ProjectWithStats
	err := h.db.Pool.QueryRow(context.Background(), `
		SELECT p.id, p.name, p.description, p.created_at, p.updated_at, COUNT(d.id)
		FROM projects p LEFT JOIN devices d ON d.project_id = p.id
		WHERE p.id = $1 GROUP BY p.id`, id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt, &p.DeviceCount)
	if err != nil {
		return errResponse(c, 404, "project not found", "NOT_FOUND")
	}
	return c.JSON(p)
}

func (h *projectHandlers) update(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errResponse(c, 400, "invalid request body", "INVALID_INPUT")
	}

	var p models.Project
	err := h.db.Pool.QueryRow(context.Background(), `
		UPDATE projects SET
			name = COALESCE(NULLIF($2, ''), name),
			description = COALESCE(NULLIF($3, ''), description),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at`,
		id, body.Name, body.Description,
	).Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return errResponse(c, 404, "project not found", "NOT_FOUND")
	}
	return c.JSON(p)
}

func (h *projectHandlers) delete(c *fiber.Ctx) error {
	id := c.Params("id")
	cmd, err := h.db.Pool.Exec(context.Background(), `DELETE FROM projects WHERE id = $1`, id)
	if err != nil || cmd.RowsAffected() == 0 {
		return errResponse(c, 404, "project not found", "NOT_FOUND")
	}
	return c.SendStatus(204)
}
