package server_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	address   string
	apiPrefix = "/v1/swift-codes"
)

func TestAddCode(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		status int
	}{
		{
			name: "valid headquarter code",
			input: `{
				"bankName": "Test Bank",
				"address": "123 Test Street",
				"countryISO2": "US",
				"countryName": "United States",
				"isHeadquarter": true,
				"swiftCode": "TESTUS33XXX"
			}`,
			status: http.StatusCreated,
			output: `{"message":"Created"}`,
		},
		{
			name: "valid branch code",
			input: `{
				"bankName": "Test Bank",
				"address": "123 Test Street",
				"countryISO2": "US",
				"countryName": "United States",
				"isHeadquarter": false,
				"swiftCode": "TESTUS33ABC"
			}`,
			status: http.StatusCreated,
			output: `{"message":"Created"}`,
		},
		{
			name: "valid branch code no headquarter",
			input: `{
				"bankName": "Test Bank",
				"address": "123 Test Street",
				"countryISO2": "PL",
				"countryName": "POLAND",
				"isHeadquarter": false,
				"swiftCode": "TESTPL33ABC"
			}`,
			status: http.StatusCreated,
			output: `{"message":"Created"}`,
		},
		{
			name: "valid headquarter code no branches",
			input: `{
				"bankName": "Test Bank",
				"address": "123 Test Street",
				"countryISO2": "US",
				"countryName": "United States",
				"isHeadquarter": true,
				"swiftCode": "TESTUS23XXX"
			}`,
			status: http.StatusCreated,
			output: `{"message":"Created"}`,
		},
		{
			name: "duplicate code",
			input: `{
				"bankName": "Test Bank",
				"address": "123 Test Street",
				"countryISO2": "US",
				"countryName": "United States",
				"isHeadquarter": true,
				"swiftCode": "TESTUS33XXX"
			}`,
			status: http.StatusConflict,
			output: `{"message":"Swift code already exists"}`,
		},
		{
			name: "invalid code",
			input: `{
				"bankName": "Test Bank",
				"address": "123 Test Street",
				"countryISO2": "US",
				"countryName": "United States",
				"isHeadquarter": true,
				"swiftCode": "INVALID"
			}`,
			status: http.StatusBadRequest,
			output: `{"message":"Validation Error: swiftCode is invalid"}`,
		},
		{
			name:   "invalid json",
			input:  `{`,
			status: http.StatusBadRequest,
			output: `{"message":"unexpected EOF"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(address+apiPrefix, "application/json", strings.NewReader(tc.input))
			assert.NoError(t, err, "failed to send request")
			defer resp.Body.Close()

			assert.Equal(t, tc.status, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "failed to read response body")

			body := strings.Trim(string(bodyBytes), "\n")
			assert.Equal(t, tc.output, string(body))
		})
	}
}

func TestDeleteCode(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		status int
	}{
		{
			name:   "existing code",
			input:  "/TESTPL33ABC",
			status: http.StatusOK,
			output: `{"message":"OK"}`,
		},
		{
			name:   "already deleted code",
			input:  "/TESTPL33ABC",
			status: http.StatusNotFound,
			output: `{"message":"Not Found"}`,
		},
		{
			name:   "nonexistent code",
			input:  "/NONEXISTENT",
			status: http.StatusNotFound,
			output: `{"message":"Not Found"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, address+apiPrefix+tc.input, nil)
			assert.NoError(t, err, "failed to create request")

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err, "failed to send request")
			defer resp.Body.Close()

			assert.Equal(t, tc.status, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "failed to read response body")

			body := strings.Trim(string(bodyBytes), "\n")
			assert.Equal(t, tc.output, body)
		})
	}
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		status int
	}{
		{
			name:   "valid branch",
			input:  "/TESTUS33ABC",
			status: http.StatusOK,
			output: `{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":false,"swiftCode":"TESTUS33ABC"}`,
		},
		{
			name:   "valid headquarter",
			input:  "/TESTUS33XXX",
			status: http.StatusOK,
			output: `{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTUS33XXX","branches":[{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","isHeadquarter":false,"swiftCode":"TESTUS33ABC"}]}`,
		},
		{
			name:   "valid headquarter no branches",
			input:  "/TESTUS23XXX",
			status: http.StatusOK,
			output: `{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","countryName":"UNITED STATES","isHeadquarter":true,"swiftCode":"TESTUS23XXX","branches":[]}`,
		},
		{
			name:   "nonexistent code",
			input:  "/NONEXISTENT",
			status: http.StatusNotFound,
			output: `{"message":"Not Found"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(address + apiPrefix + tc.input)
			assert.NoError(t, err, "failed to send request")
			defer resp.Body.Close()

			assert.Equal(t, tc.status, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "failed to read response body")

			body := strings.Trim(string(bodyBytes), "\n")
			assert.Equal(t, tc.output, string(body))
		})
	}
}

func TestGetByCountryCode(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
		status int
	}{
		{
			name:   "valid country",
			input:  "US",
			status: http.StatusOK,
			output: `{"countryISO2":"US","countryName":"UNITED STATES","swiftCodes":[{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","isHeadquarter":true,"swiftCode":"TESTUS33XXX"},{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","isHeadquarter":false,"swiftCode":"TESTUS33ABC"},{"address":"123 Test Street","bankName":"Test Bank","countryISO2":"US","isHeadquarter":true,"swiftCode":"TESTUS23XXX"}]}`,
		},
		{
			name:   "nonexistent country",
			input:  "XX",
			status: http.StatusNotFound,
			output: `{"message":"Not Found"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Get(address + apiPrefix + "/country/" + tc.input)
			assert.NoError(t, err, "failed to send request")
			defer resp.Body.Close()

			assert.Equal(t, tc.status, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, "failed to read response body")

			body := strings.Trim(string(bodyBytes), "\n")
			assert.Equal(t, tc.output, string(body))
		})
	}
}

func TestMain(m *testing.M) {
	c := context.Background()
	stack, err := compose.NewDockerCompose("../compose.yaml")
	if err != nil {
		log.Fatalf("Failed to create stack: %v\n", err)
	}

	err = stack.WithEnv(map[string]string{
		"DATABASE_USERNAME": "test",
		"DATABASE_PASSWORD": "test",
		"DATABASE_NAME":     "test",
		"API_PORT":          "",
	}).WaitForService("api", wait.ForListeningPort("8080/tcp")).Up(c, compose.Wait(true))
	if err != nil {
		log.Fatalf("Failed to start stack: %v\n", err)
	}
	defer func() {
		err = stack.Down(
			context.Background(),
			compose.RemoveOrphans(true),
			compose.RemoveVolumes(true),
			compose.RemoveImagesLocal,
		)
		if err != nil {
			log.Printf("Failed to stop stack: %v\n", err)
		}
	}()

	apiContainer, err := stack.ServiceContainer(context.Background(), "api")
	if err != nil {
		log.Fatalf("Failed to get container: %v\n", err)
	}
	host, err := apiContainer.Host(c)
	if err != nil {
		log.Fatalf("Failed to get host: %v\n", err)
	}
	port, err := apiContainer.MappedPort(c, "8080/tcp")
	if err != nil {
		log.Fatalf("Failed to get port: %v\n", err)
	}
	address = fmt.Sprintf("http://%s:%s", host, port.Port())

	exitCode := m.Run()
	os.Exit(exitCode)
}
