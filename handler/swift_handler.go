package handler

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/rtsncs/remitly-swift-api/models"
)

type responseWithBranches struct {
	models.SwiftCode
	Branches []models.SwiftCode `json:"branches"`
}

type responseByCountry struct {
	CountryISO2 string             `json:"countryISO2"`
	CountryName string             `json:"countryName"`
	SwiftCodes  []models.SwiftCode `json:"swiftCodes"`
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

		return c.JSON(http.StatusOK, responseWithBranches{SwiftCode: codeDetails, Branches: branches})
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

func (h *Handler) AddCode(c echo.Context) error {
	code := new(models.SwiftCode)
	if err := c.Bind(code); err != nil {
		return err
	}
	if err := code.Validate(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if err := h.db.InsertCode(c.Request().Context(), *code); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return echo.NewHTTPError(http.StatusConflict, "Swift code already exists")
		}
		return err
	}

	return c.JSON(http.StatusCreated, genericResponse{http.StatusText(http.StatusCreated)})
}

func (h *Handler) DeleteCode(c echo.Context) error {
	code := c.Param("code")

	count, err := h.db.DeleteByCode(c.Request().Context(), code)
	if err != nil {
		return err
	}
	if count == 0 {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, genericResponse{http.StatusText(http.StatusOK)})
}
