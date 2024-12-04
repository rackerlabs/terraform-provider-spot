package provider

import (
	"context"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
)

const (
	keyResourceVersion = "resource_version"

	// attribute names defined in the provider_code_spec.json are
	// defined as constants here, to avoid typos.
	// Make sure to update these if the provider_code_spec.json changes.
	attribRegion         = "region"
	attribServerClass    = "server_class"
	attribDeploymentType = "deployment_type"
)

func listRegions(ctx context.Context, client ngpc.Client) ([]ngpcv1.Region, error) {
	regionsList := ngpcv1.RegionList{}
	err := client.List(ctx, &regionsList)
	if err != nil {
		return nil, err
	}
	return regionsList.Items, nil
}

func listServerClasses(ctx context.Context, client ngpc.Client) ([]ngpcv1.ServerClass, error) {
	var serverClassList ngpcv1.ServerClassList
	err := client.List(ctx, &serverClassList)
	if err != nil {
		return nil, err
	}
	return serverClassList.Items, nil
}
