/*
 * Copyright 2019 InfAI (CC SES)
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
	"log"
	"os"

	"github.com/SENERGY-Platform/analytics-fog-master/lib"
	"github.com/SENERGY-Platform/analytics-fog-master/lib/config"
	"github.com/joho/godotenv"
)
 
 func main() {
	 ec := 0
	 defer func() {
		 os.Exit(ec)
	 }()
 
	 err := godotenv.Load()
	 if err != nil {
		 log.Print("Error loading .env file: ", err)
		 ec = 1
		 return
	 }
 
	 config, err := config.NewConfig("")
	 if err != nil {
		 log.Print("Cant load config: ", err)
		 ec = 1
		 return
	 }
	 
	 ctx := context.Background()
	 err = lib.Run(ctx, os.Stdout, os.Stderr, *config)
	 if err != nil {
		 ec = 1
		 return
	 }
 }
 