package utils

import (
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"context"
	"os"
	"path/filepath"
	"fmt"
)


type Mosquitto struct {
	container testcontainers.Container;
}

func NewMosquitto(ctx context.Context) (*Mosquitto, error) {
	absPath, err := filepath.Abs(filepath.Join("..", "utils", "conf.conf"))
	if err != nil {
		return &Mosquitto{}, err
	}
	r, err := os.Open(absPath)
	if err != nil {
		return &Mosquitto{}, err 
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:           "eclipse-mosquitto:2.0.18",
			Tmpfs:           map[string]string{},
			ExposedPorts:    []string{"1883/tcp"},
			WaitingFor:      wait.ForListeningPort("1883/tcp"),
			AlwaysPullImage: true,
			Files: []testcontainers.ContainerFile{
				{
					Reader:            r,
					HostFilePath:      "./tests/utils/conf.conf", // will be discarded internally
					ContainerFilePath: "/mosquitto/config/mosquitto.conf",
					FileMode:          0o777,
				},
			},
		},
		Started: false,
	})
	if err != nil {
		return &Mosquitto{}, err
	}
	return &Mosquitto{
		container: container,
	}, nil
}

func (m *Mosquitto) StartAndWait(ctx context.Context) (error, string) {
	err := m.container.Start(ctx)
	if err != nil {
		return err, ""
	}
	localhostPort, err := m.container.MappedPort(ctx, "1883")
	fmt.Println("Exposed broker at: ", localhostPort.Port())
	if err != nil {
		return err, ""
	}
	return nil, localhostPort.Port()
}