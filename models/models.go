package models

import (
	"regexp"
	"strings"
)

var (
	swiftCodeRegex   = regexp.MustCompile(`^[A-Z]{6}[A-Z0-9]{5}$`)
	countryCodeRegex = regexp.MustCompile(`^[A-Z]{2}$`)
)

type FieldError struct {
	Name    string `json:"name"`
	Details string `json:"details"`
}

type FieldErrors []FieldError

func (fe FieldErrors) Error() string {
	var b []byte
	b = append(b, "Validation Error: "...)
	for _, e := range fe {
		b = append(b, e.Name...)
		b = append(b, " "...)
		b = append(b, e.Details...)
		b = append(b, "; "...)
	}

	return strings.TrimSuffix(string(b), "; ")
}

type SwiftCode struct {
	Address     string `json:"address"`
	Name        string `json:"bankName"`
	CountryISO2 string `json:"countryISO2"`
	CountryName string `json:"countryName,omitempty"`
	Headquarter bool   `json:"isHeadquarter"`
	Code        string `json:"swiftCode"`
}

func (code *SwiftCode) Validate() error {
	var fe FieldErrors

	code.Code = strings.ToUpper(code.Code)
	code.CountryISO2 = strings.ToUpper(code.CountryISO2)
	code.CountryName = strings.ToUpper(code.CountryName)

	if code.Name == "" {
		fe = append(fe, FieldError{"bankName", "is required"})
	}

	if code.CountryISO2 == "" {
		fe = append(fe, FieldError{"countryISO2", "is required"})
	} else if !countryCodeRegex.MatchString(code.CountryISO2) {
		fe = append(fe, FieldError{"countryISO2", "must consist of two ASCII letters"})
	}

	if code.CountryName == "" {
		fe = append(fe, FieldError{"countryName", "is required"})
	}

	if code.Code == "" {
		fe = append(fe, FieldError{"swiftCode", "is required"})
	} else if !swiftCodeRegex.MatchString(code.Code) {
		fe = append(fe, FieldError{"swiftCode", "is invalid"})
	} else {
		if code.Code[4:6] != code.CountryISO2 {
			fe = append(fe, FieldError{"swiftCode", "doesn't match countryISO2"})
		}
		if (code.Code[8:] == "XXX") != code.Headquarter {
			fe = append(fe, FieldError{"isHeadquarter", "doesn't match swiftCode"})
		}
	}

	if len(fe) > 0 {
		return fe
	}

	return nil
}
