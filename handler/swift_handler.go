package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rtsncs/remitly-swift-api/database"
)

func (h *Handler) GetCode(c echo.Context) error {
	code := c.Param("code")

	codeDetails, err := h.db.GetByCode(c.Request().Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}
	if codeDetails.Headquarter {
		branches, err := h.db.GetBranches(c.Request().Context(), code)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}

		return c.JSON(http.StatusOK, database.SwiftCodeWithBranches{SwiftCode: codeDetails, Branches: branches})
	}

	return c.JSON(http.StatusOK, codeDetails)
}
