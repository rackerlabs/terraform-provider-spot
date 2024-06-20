package provider

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	DefaultCloudSpaceCreateTimeout = 30 * time.Minute
	DefaultCloudSpaceUpdateTimeout = 20 * time.Minute
)

// This function returns retry configuration for cloudspace to be ready
func getCloudSpaceReadyRetryConfig() *backoff.ExponentialBackOff {
	exponentialBackoff := backoff.NewExponentialBackOff()
	// Will return error after MaxElapsedTime
	exponentialBackoff.MaxElapsedTime = DefaultCloudSpaceCreateTimeout
	// first retry attempt will be made after InitialInterval
	exponentialBackoff.InitialInterval = 5 * time.Minute
	// max delay between retries, delay between retries will vary using randomization factor
	exponentialBackoff.MaxInterval = 2 * time.Minute
	return exponentialBackoff
}
