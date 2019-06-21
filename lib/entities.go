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

package lib

import "time"

type ControlCommand struct {
	Command string      `json:"command,omitempty"`
	Data    OperatorJob `json:"data,omitempty"`
}

type OperatorJob struct {
	PipelineId  string     `json:"pipelineId,omitempty"`
	OperatorId  string     `json:"operatorId,omitempty"`
	ImageId     string     `json:"imageId,omitempty"`
	Agent       Agent      `json:"agent,omitempty"`
	ContainerId string     `json:"containerId,omitempty"`
	Config      []struct{} `json:"config,omitempty"`
}

type AgentMessage struct {
	Type string `json:"type,omitempty"`
	Conf Agent  `json:"agent,omitempty"`
}

type Agent struct {
	Id      string    `json:"id,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
	Active  bool      `json:"active,omitempty"`
}