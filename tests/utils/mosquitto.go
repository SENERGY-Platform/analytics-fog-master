/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type Mosquitto struct {
	container testcontainers.Container
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
