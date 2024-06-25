package provider

import (
	"time"
)

const (
	// DefaultCloudSpaceCreateTimeout is the default timeout for creating a control plane for a cloud space.
	DefaultCloudSpaceCreateTimeout = 5 * time.Minute
	// DefaultRefreshInterval is the default interval at which the provider will poll the API for updates.
	DefaultRefreshInterval = 5 * time.Second
)
