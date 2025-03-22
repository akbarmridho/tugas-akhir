package health

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"tugas-akhir/backend/infrastructure/postgres"
	myerror "tugas-akhir/backend/pkg/error"
)

type PGHealthcheckHandler struct {
	db *postgres.Postgres
}

func NewPGHealthcheckHandler(db *postgres.Postgres) *PGHealthcheckHandler {
	return &PGHealthcheckHandler{
		db: db,
	}
}

func (h *PGHealthcheckHandler) Healthcheck(c echo.Context) error {
	err := h.db.Pool.Ping(c.Request().Context())

	if err != nil {
		return c.JSON(http.StatusInternalServerError, myerror.HttpPayload{
			Message: "Failed to ping database",
		})
	}

	return c.JSON(http.StatusOK, myerror.HttpPayload{
		Message: "Ok",
	})
}
