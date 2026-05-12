/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/SENERGY-Platform/analytics-fog-master/internal/config"
	"github.com/SENERGY-Platform/analytics-fog-master/lib"
)

func main() {
	ec := 0
	defer func() {
		os.Exit(ec)
	}()

	config.ParseFlags()

	cfg, err := config.New(config.ConfPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		ec = 1
		return
	}

	ctx := context.Background()
	err = lib.Run(ctx, os.Stdout, os.Stderr, *cfg)
	if err != nil {
		log.Printf("Error running: %s", err)
		ec = 1
		return
	}
}
