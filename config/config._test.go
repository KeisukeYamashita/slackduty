package config

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func sweepEnvs() {
	os.Setenv("SLACKDUTY_CONFIG", "")
	os.Setenv("SLACKDUTY_PAGERDUTY_API_KEY", "")
	os.Setenv("SLACKDUTY_SLACK_API_KEY", "")
	os.Setenv("SLACKDUTY_EXTENRAL_TRIGGER", "")
}

func TestLoad(t *testing.T) {
	testGroups := []Group{
		Group{
			Name:       "Slackduty on-support Slack usergroup",
			Usergroups: []string{"handle:slackduty-on-support"},
			Schedule:   "* * * * *",
			Members: &Members{
				Pagerduty: &Pagerduty{
					Teams:     []string{"name:slackdutyPrimary"},
					Services:  []string{"name:slackduty-backend"},
					Schedules: []string{"name:slackduty-oncall"},
				},
			},
			Exclude: []string{"name:slackduty@example.com"},
		},
	}

	tcs := map[string]struct {
		filepath string
		success  bool
		want     *Config
	}{
		"pass":             {"./example.yml", true, &Config{Groups: testGroups}},
		"invalid filepath": {"./invalid/filepath/for/config.bad", false, nil},
	}

	for n, tc := range tcs {
		t.Run(n, func(t *testing.T) {
			got, err := Load(tc.filepath)
			if err != nil {
				if tc.success {
					t.Fatalf("test %s error: %v", n, err)
				} else {
					return
				}
			}

			if !reflect.DeepEqual(got, tc.want) {
				diff := cmp.Diff(got, tc.want)
				t.Fatalf("load result unexpected diff:%v", diff)
			}
		})
	}
}

func TestGetConfigPath(t *testing.T) {
	tcs := map[string]struct {
		fn   func()
		want string
	}{
		"default":   {func() {}, defaultConfigPath},
		"conifgure": {func() { os.Setenv("SLACKDUTY_CONFIG", "~/slackduty/test/config.yml") }, "~/slackduty/test/config.yml"},
	}

	for n, tc := range tcs {
		tc.fn()
		if got := GetConfigPath(); !strings.Contains(got, tc.want) {
			t.Fatalf("not expected %s got: %s want: %s", n, got, tc.want)
		}
		sweepEnvs()
	}
}

func TestGetAPIKeys(t *testing.T) {
	tcs := map[string]struct {
		fn   func()
		want struct {
			pdAPIKey    string
			slackAPIKey string
		}
		success bool
	}{
		"default": {
			fn: func() {},
			want: struct {
				pdAPIKey    string
				slackAPIKey string
			}{"", ""},
			success: false},
		"pass": {
			fn: func() {
				os.Setenv("SLACKDUTY_PAGERDUTY_API_KEY", "pd_key")
				os.Setenv("SLACKDUTY_SLACK_API_KEY", "slack_key")
			},
			want: struct {
				pdAPIKey    string
				slackAPIKey string
			}{"pb_key", "slack_key"},
			success: true},
	}

	for n, tc := range tcs {
		tc.fn()
		pdAPIKey, slackAPIKey, err := GetAPIKeys()

		if err != nil {
			if tc.success {
				t.Fatalf("test %s error: %v", n, err)
			} else {
				return
			}
		}

		if pdAPIKey != tc.want.pdAPIKey && slackAPIKey != tc.want.slackAPIKey {
			t.Fatalf("not expected API keys %s got pdAPIKey:%s slackAPIKey:%s, want: pdAPIKey:%s slackAPIKey:%s", n, pdAPIKey, tc.want.pdAPIKey, slackAPIKey, tc.want.slackAPIKey)
		}
		sweepEnvs()
	}
}

func TestIsExternalTrigger(t *testing.T) {
	tcs := map[string]struct {
		fn   func()
		want bool
	}{
		"default":   {func() {}, false},
		"conifgure": {func() { os.Setenv("SLACKDUTY_EXTENRAL_TRIGGER", "true") }, true},
	}

	for n, tc := range tcs {
		tc.fn()
		if got := IsExternalTrigger(); got != tc.want {
			t.Fatalf("not expected %s got: %v want: %v", n, got, tc.want)
		}
		sweepEnvs()
	}
}
