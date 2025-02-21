package provider

import (
	"context"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SpotProviderClient struct {
	*ngpc.HTTPClient
	organizerConfig *rest.Config
	organizer       *ngpc.OrganizerClient
}

func NewSpotProviderClient(client *ngpc.HTTPClient) *SpotProviderClient {
	return &SpotProviderClient{
		HTTPClient:      client,
		organizerConfig: client.Config,
		organizer:       nil,
	}
}

func (c *SpotProviderClient) GetOrganizer() *ngpc.OrganizerClient {
	if c.organizer == nil {
		c.organizer = ngpc.NewOrganizerClient(c.organizerConfig)
	}
	return c.organizer
}

// Get implements the client.Get method
func (c *SpotProviderClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object) error {
	return c.HTTPClient.Get(ctx, key, obj)
}

// Create implements the client.Create method
func (c *SpotProviderClient) Create(ctx context.Context, obj client.Object) error {
	return c.HTTPClient.Create(ctx, obj)
}

// Update implements the client.Update method
func (c *SpotProviderClient) Update(ctx context.Context, obj client.Object) error {
	return c.HTTPClient.Update(ctx, obj)
}

// Delete implements the client.Delete method
func (c *SpotProviderClient) Delete(ctx context.Context, obj client.Object) error {
	return c.HTTPClient.Delete(ctx, obj)
}
