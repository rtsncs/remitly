package server_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	address   string
	apiPrefix = "/v1/swift-codes"
)

func TestGetCode(t *testing.T) {
	resp, err := http.Get(address + apiPrefix + "/NONEXISTENT")
	if err != nil {
		t.Errorf("Response error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got: %d\n", resp.StatusCode)
	}
}

func TestGetByCountryCode(t *testing.T) {
	resp, err := http.Get(address + apiPrefix + "/countries/XX")
	if err != nil {
		t.Errorf("Response error: %v\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got: %d\n", resp.StatusCode)
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
