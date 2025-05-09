package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSwiftCode(t *testing.T) {
	tests := []struct {
		name     string
		input    SwiftCode
		wantErr  bool
		wantMsgs []string
	}{
		{
			name: "valid headquarter code",
			input: SwiftCode{
				Address:       "Some Street 123",
				BankName:      "Bank of Test",
				CountryISO2:   "PL",
				CountryName:   "Poland",
				IsHeadquarter: true,
				SwiftCode:     "BANKPLPWXXX",
			},
			wantErr: false,
		},
		{
			name: "valid branch code",
			input: SwiftCode{
				Address:       "Branch Road 456",
				BankName:      "Bank of Test",
				CountryISO2:   "PL",
				CountryName:   "Poland",
				IsHeadquarter: false,
				SwiftCode:     "BANKPLPW001",
			},
			wantErr: false,
		},
		{
			name:     "missing required fields",
			input:    SwiftCode{},
			wantErr:  true,
			wantMsgs: []string{"bankName", "countryISO2", "countryName", "swiftCode"},
		},
		{
			name: "invalid country code format",
			input: SwiftCode{
				BankName:      "Bank",
				CountryISO2:   "PL1",
				CountryName:   "Poland",
				SwiftCode:     "BANKPLPWXXX",
				IsHeadquarter: true,
			},
			wantErr:  true,
			wantMsgs: []string{"countryISO2"},
		},
		{
			name: "invalid swift code format",
			input: SwiftCode{
				BankName:      "Bank",
				CountryISO2:   "PL",
				CountryName:   "Poland",
				SwiftCode:     "INVALID",
				IsHeadquarter: true,
			},
			wantErr:  true,
			wantMsgs: []string{"swiftCode"},
		},
		{
			name: "country code mismatch in swift",
			input: SwiftCode{
				BankName:      "Bank",
				CountryISO2:   "DE",
				CountryName:   "Germany",
				SwiftCode:     "BANKPLPWXXX",
				IsHeadquarter: true,
			},
			wantErr:  true,
			wantMsgs: []string{"swiftCode"},
		},
		{
			name: "headquarter mismatch",
			input: SwiftCode{
				BankName:      "Bank",
				CountryISO2:   "PL",
				CountryName:   "Poland",
				SwiftCode:     "BANKPLPWXXX",
				IsHeadquarter: false,
			},
			wantErr:  true,
			wantMsgs: []string{"isHeadquarter"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.wantErr {
				assert.Error(t, err)
				var fe FieldErrors
				assert.ErrorAs(t, err, &fe)
				for _, msg := range tc.wantMsgs {
					found := false
					for _, field := range fe {
						if strings.Contains(field.Name, msg) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error containing field '%s', but it was missing in: %v", msg, fe)
					}
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
