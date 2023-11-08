/*
Copyright 2016 Skippbox, Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"github.com/marvasgit/kubernetes-diffwatcher/config"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/event"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/flock"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/hipchat"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/lark"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/mattermost"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/msteam"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/slack"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/slackwebhook"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/smtpClient"
	"github.com/marvasgit/kubernetes-diffwatcher/pkg/handlers/webhook"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Init(c *config.Config) error
	Handle(e event.DiffWatchEvent)
}

// Map maps each event handler function to a name for easily lookup
var Map = map[string]interface{}{
	"default":      &Default{},
	"slack":        &slack.Slack{},
	"slackwebhook": &slackwebhook.SlackWebhook{},
	"hipchat":      &hipchat.Hipchat{},
	"mattermost":   &mattermost.Mattermost{},
	"flock":        &flock.Flock{},
	"webhook":      &webhook.Webhook{},
	"ms-teams":     &msteam.MSTeams{},
	"smtp":         &smtpClient.SMTP{},
	"lark":         &lark.Webhook{},
}

// Default handler implements Handler interface,
// print each event with JSON format
type Default struct {
}

// Init initializes handler configuration
// Do nothing for default handler
func (d *Default) Init(c *config.Config) error {
	return nil
}

// Handle handles an event.
func (d *Default) Handle(e event.DiffWatchEvent) {
}
