package client

import (
	"testing"

	"github.com/KeisukeYamashita/slackduty/config"
	"github.com/KeisukeYamashita/slackduty/log"
)

func TestWithExtenralTrigger(t *testing.T) {
	tcs := map[string]struct {
		trigger bool
	}{
		"external trigger": {true},
	}

	for n := range tcs {
		t.Run(n, func(t *testing.T) {
			opt := WithExternalTrigger()
			var o options
			opt(&o)

			if !o.externalTrigger {
				t.Fatal("external trigger should be set to true")
			}
		})
	}
}

func TestNew(t *testing.T) {
	const (
		testPdAPIKey    = "test-pd-key"
		testSlackAPIKey = "test-slack-key"
	)

	testLogger := log.NewDiscard()

	type input struct {
		config      *config.Config
		pdAPIKey    string
		slackAPIKey string
	}

	tcs := map[string]struct {
		input *input
		want  *Client
	}{
		"api keys configured": {&input{&config.Config{}, testPdAPIKey, testSlackAPIKey}, &Client{pagerduty: NewPagerDutyClient(testPdAPIKey), slack: NewSlackClient(testSlackAPIKey), logger: testLogger}},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			got := New(tc.input.config, tc.input.pdAPIKey, tc.input.slackAPIKey, testLogger)
			_ = got
		})
	}
}
