package main

import "strings"

// verbs are the operationId words that make up a command's verb phrase.
// The phrase is the maximal leading run of them, so bulkDeleteValidate
// stays one verb rather than a verb plus two nouns.
var verbs = map[string]bool{
	"add": true, "bulk": true, "cancel": true, "comment": true, "copy": true,
	"create": true, "delete": true, "generate": true, "get": true, "list": true,
	"mute": true, "query": true, "redeliver": true, "resend": true, "reset": true,
	"resolve": true, "resume": true, "set": true, "test": true, "update": true,
	"validate": true, "verify": true, "write": true,
}

// entityOverrides names the noun a tag's operationIds actually use, where
// it is not the tag's own singular. MonitoringLocations publishes agents.
var entityOverrides = map[string][]string{
	"MonitoringLocations": {"agent"},
}

// nameOverrides replaces a mechanical command name that reads worse than
// the operation it stands for.
var nameOverrides = map[string]string{
	"redeliverWebhookDelivery": "redeliver",
}

// groupShort is the one-line help of each group. The document publishes
// no tag descriptions, so the CLI carries its own.
var groupShort = map[string]string{
	"Monitors":            "Create, read, change and bulk-edit monitors",
	"MonitorTypes":        "The catalogue of check types and their settings schema",
	"Results":             "Individual check results and their summaries",
	"Incidents":           "Downtime incidents, their checks and their comments",
	"Maintenance":         "Planned maintenance windows",
	"Contacts":            "Notification contacts, contact groups and confirmations",
	"Alerts":              "Who is alerted about which monitor, and what was sent",
	"Reports":             "Report subscriptions and generated reports",
	"Webhooks":            "Webhook endpoints, their deliveries and test sends",
	"StatusPages":         "Public status pages, their incidents, templates and subscribers",
	"Account":             "The account, its quota and its usage",
	"MonitoringLocations": "The monitoring locations checks can run from",
	"InstantChecks":       "One-off checks run on demand, without a monitor",
	"Jobs":                "Long-running batch jobs started by the bulk operations",
}

// groupAliases are extra spellings of a group name.
var groupAliases = map[string][]string{
	"instant-checks":       {"checks"},
	"monitoring-locations": {"locations", "agents"},
	"status-pages":         {"statuspages"},
	"monitor-types":        {"types"},
}

// words splits a camelCase identifier into lower-case words. A run of
// capitals is one word up to the last of them, so listAgentIp yields
// list, agent, ip and getHTTPStatus would yield get, http, status.
func words(identifier string) []string {
	runes := []rune(identifier)
	var out []string
	start := 0
	flush := func(end int) {
		if end > start {
			out = append(out, strings.ToLower(string(runes[start:end])))
		}
	}
	for i := 0; i < len(runes); i++ {
		switch {
		case runes[i] == '_' || runes[i] == '-' || runes[i] == '.' || runes[i] == ' ':
			flush(i)
			start = i + 1
		case i > start && isUpper(runes[i]) && !isUpper(runes[i-1]):
			flush(i)
			start = i
		case i > start && isUpper(runes[i]) && isUpper(runes[i-1]) && i+1 < len(runes) && isLower(runes[i+1]):
			flush(i)
			start = i
		}
	}
	flush(len(runes))
	return out
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }

// singular is a naive singulariser, enough for the nouns the tags use.
// A word ending in "us" or "ss" is left alone: status is not statu.
func singular(word string) string {
	switch {
	case strings.HasSuffix(word, "ies") && len(word) > 3:
		return word[:len(word)-3] + "y"
	case strings.HasSuffix(word, "ss"), strings.HasSuffix(word, "us"):
		return word
	case strings.HasSuffix(word, "ses"):
		return word[:len(word)-2]
	case strings.HasSuffix(word, "s"):
		return word[:len(word)-1]
	default:
		return word
	}
}

// groupName spells a tag as its command group: the tag in lower kebab,
// with the words the document published.
func groupName(tag string) string {
	return strings.Join(words(tag), "-")
}

// entityWords are the nouns a group's commands drop, because the group
// name already says them.
func entityWords(tag string) map[string]bool {
	out := map[string]bool{}
	if override, ok := entityOverrides[tag]; ok {
		for _, word := range override {
			out[word] = true
		}
		return out
	}
	for _, word := range words(tag) {
		out[singular(word)] = true
	}
	return out
}

// commandName maps an operationId onto a command name inside its group:
// the verb phrase, followed by whatever qualifies it once the group's own
// noun is dropped.
//
//	listMonitor          -> list
//	getMonitorAttached   -> get-attached
//	bulkDeleteValidate.. -> bulk-delete-validate
//	deleteContactAlerts  -> delete-contact-alerts
//
// A plural of the group's noun is kept, which is what separates the
// operation on one subscription (delete-contact) from the one on all of a
// contact's (delete-contact-alerts).
func commandName(operationID, tag string) string {
	if override, ok := nameOverrides[operationID]; ok {
		return override
	}
	entity := entityWords(tag)
	parts := words(operationID)

	verbEnd := 0
	for verbEnd < len(parts) && verbs[parts[verbEnd]] {
		verbEnd++
	}
	name := append([]string(nil), parts[:verbEnd]...)
	for _, word := range parts[verbEnd:] {
		if entity[word] {
			continue
		}
		name = append(name, word)
	}
	return strings.Join(name, "-")
}

// flagName spells a parameter name as a command-line flag: camelCase,
// underscores and the dots of a scoped filter become dashes, and a name
// that already carries one keeps exactly one (Idempotency-Key is
// idempotency-key, not idempotency--key; monitor.id is monitor-id).
func flagName(name string) string {
	var b strings.Builder
	dashed := func() bool { return b.Len() == 0 || strings.HasSuffix(b.String(), "-") }
	for i, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			if i > 0 && !dashed() {
				b.WriteByte('-')
			}
			b.WriteRune(r - 'A' + 'a')
		case r == '_', r == '-', r == '.':
			if !dashed() {
				b.WriteByte('-')
			}
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// fileName is the generated file a group lands in.
func fileName(group string) string {
	return strings.ReplaceAll(group, "-", "_") + ".go"
}
