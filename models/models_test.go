package models

import (
	"strings"
	"testing"
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
				Address:     "Some Street 123",
				Name:        "Bank of Test",
				CountryISO2: "PL",
				CountryName: "Poland",
				Headquarter: true,
				Code:        "BANKPLPWXXX",
			},
			wantErr: false,
		},
		{
			name: "valid branch code",
			input: SwiftCode{
				Address:     "Branch Road 456",
				Name:        "Bank of Test",
				CountryISO2: "PL",
				CountryName: "Poland",
				Headquarter: false,
				Code:        "BANKPLPW001",
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
				Name:        "Bank",
				CountryISO2: "PL1",
				CountryName: "Poland",
				Code:        "BANKPLPWXXX",
				Headquarter: true,
			},
			wantErr:  true,
			wantMsgs: []string{"countryISO2"},
		},
		{
			name: "invalid swift code format",
			input: SwiftCode{
				Name:        "Bank",
				CountryISO2: "PL",
				CountryName: "Poland",
				Code:        "INVALID",
				Headquarter: true,
			},
			wantErr:  true,
			wantMsgs: []string{"swiftCode"},
		},
		{
			name: "country code mismatch in swift",
			input: SwiftCode{
				Name:        "Bank",
				CountryISO2: "DE",
				CountryName: "Germany",
				Code:        "BANKPLPWXXX",
				Headquarter: true,
			},
			wantErr:  true,
			wantMsgs: []string{"swiftCode"},
		},
		{
			name: "headquarter mismatch",
			input: SwiftCode{
				Name:        "Bank",
				CountryISO2: "PL",
				CountryName: "Poland",
				Code:        "BANKPLPWXXX",
				Headquarter: false,
			},
			wantErr:  true,
			wantMsgs: []string{"isHeadquarter"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.wantErr && err == nil {
				t.Errorf("Expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if fe, ok := err.(FieldErrors); ok {
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
			}
		})
	}
}
