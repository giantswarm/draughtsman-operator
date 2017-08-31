package notifier

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/notifier/slack"
)

type Notifier struct {
	Slack slack.Slack
	Type  string
}
