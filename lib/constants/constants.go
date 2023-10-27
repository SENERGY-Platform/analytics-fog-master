/*
 * Copyright 2020 InfAI (CC SES)
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

package constants

import (
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/agent"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/control"
	"github.com/SENERGY-Platform/analytics-fog-lib/lib/operator"
)

const AgentsTopic = agent.AgentsTopic

// TODO control topic per user id control/[USERID] this hsould be done at the connector though
// login at keycloak to get user id
const ControlTopic = control.ControlTopic
const OperatorsTopic = operator.OperatorsTopic
