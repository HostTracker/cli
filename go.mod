module github.com/HostTracker/cli

go 1.24.0

require (
	github.com/HostTracker/hosttracker-sdk-go v0.1.0
	github.com/google/uuid v1.6.0
	github.com/spf13/cobra v1.10.1
	github.com/spf13/pflag v1.0.10
	golang.org/x/term v0.36.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/oapi-codegen/runtime v1.7.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
)

// The SDK is not published yet. Dropped at release, when the tagged
// module replaces the sibling checkout.
replace github.com/HostTracker/hosttracker-sdk-go => ../hosttracker-sdk-go
