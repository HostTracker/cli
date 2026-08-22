package main

import (
	"reflect"
	"testing"
)

func TestWords(t *testing.T) {
	cases := map[string][]string{
		"listMonitor":                        {"list", "monitor"},
		"bulkDeleteValidateMonitor":          {"bulk", "delete", "validate", "monitor"},
		"listAgentIp":                        {"list", "agent", "ip"},
		"getMonitorSettingsSchema":           {"get", "monitor", "settings", "schema"},
		"addStatusPageIncidentTimelineEntry": {"add", "status", "page", "incident", "timeline", "entry"},
		"StatusPages":                        {"status", "pages"},
		"MonitoringLocations":                {"monitoring", "locations"},
		"Idempotency-Key":                    {"idempotency", "key"},
	}
	for input, want := range cases {
		if got := words(input); !reflect.DeepEqual(got, want) {
			t.Errorf("words(%q) = %v, want %v", input, got, want)
		}
	}
}

func TestCommandName(t *testing.T) {
	cases := []struct {
		operationID string
		tag         string
		want        string
	}{
		{"listMonitor", "Monitors", "list"},
		{"getMonitor", "Monitors", "get"},
		{"bulkCreateMonitor", "Monitors", "bulk-create"},
		{"bulkDeleteValidateMonitor", "Monitors", "bulk-delete-validate"},
		{"getMonitorAttached", "Monitors", "get-attached"},
		{"resolveMonitorUrl", "Monitors", "resolve-url"},
		{"muteMonitorDnsbl", "Monitors", "mute-dnsbl"},
		{"getMonitorSettingsSchema", "MonitorTypes", "get-settings-schema"},
		{"getMonitorResultSnapshot", "Results", "get-monitor-snapshot"},
		{"listAlertByContact", "Alerts", "list-by-contact"},
		{"listContactAlert", "Alerts", "list-contact"},
		// A plural of the group's own noun is kept, which is what keeps
		// the one-subscription delete apart from the all-of-them delete.
		{"deleteContactAlert", "Alerts", "delete-contact"},
		{"deleteContactAlerts", "Alerts", "delete-contact-alerts"},
		{"createStatusPage", "StatusPages", "create"},
		{"addStatusPageSubscriber", "StatusPages", "add-subscriber"},
		{"listAgentPool", "MonitoringLocations", "list-pool"},
		{"createInstantCheck", "InstantChecks", "create"},
		{"redeliverWebhookDelivery", "Webhooks", "redeliver"},
	}
	for _, c := range cases {
		if got := commandName(c.operationID, c.tag); got != c.want {
			t.Errorf("commandName(%q, %q) = %q, want %q", c.operationID, c.tag, got, c.want)
		}
	}
}

func TestGroupName(t *testing.T) {
	cases := map[string]string{
		"Monitors":            "monitors",
		"MonitorTypes":        "monitor-types",
		"StatusPages":         "status-pages",
		"InstantChecks":       "instant-checks",
		"MonitoringLocations": "monitoring-locations",
	}
	for tag, want := range cases {
		if got := groupName(tag); got != want {
			t.Errorf("groupName(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestFlagName(t *testing.T) {
	cases := map[string]string{
		"limit":           "limit",
		"updatedSince":    "updated-since",
		"Idempotency-Key": "idempotency-key",
		"If-None-Match":   "if-none-match",
		"monitor.id":      "monitor-id",
		"openStat":        "open-stat",
	}
	for input, want := range cases {
		if got := flagName(input); got != want {
			t.Errorf("flagName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSingular(t *testing.T) {
	cases := map[string]string{
		"monitors": "monitor",
		"types":    "type",
		"pages":    "page",
		"status":   "status",
		"checks":   "check",
	}
	for input, want := range cases {
		if got := singular(input); got != want {
			t.Errorf("singular(%q) = %q, want %q", input, got, want)
		}
	}
}
