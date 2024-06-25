package provider

import (
	"time"
)

const (
	DefaultCloudSpaceCreateTimeout = 30 * time.Minute
	DefaultCloudSpaceUpdateTimeout = 20 * time.Minute
	// DefaultRefreshInterval is the default interval at which the provider will poll the API for updates.
	DefaultRefreshInterval = 10 * time.Second
)
