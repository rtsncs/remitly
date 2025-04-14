package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/rtsncs/remitly-swift-api/database"
)

type responseByCountry struct {
	CountryISO2 string               `json:"countryISO2"`
	CountryName string               `json:"countryName"`
	SwiftCodes  []database.SwiftCode `json:"swiftCodes"`
}

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

func (h *Handler) GetByCountryCode(c echo.Context) error {
	countryCode := c.Param("countryCode")

	name, err := h.db.GetCountryName(c.Request().Context(), countryCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		return err
	}

	codes, err := h.db.GetByCountryCode(c.Request().Context(), countryCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			codes = nil
		} else {
			return err
		}
	}

	response := responseByCountry{countryCode, name, codes}

	return c.JSON(http.StatusOK, response)
}
