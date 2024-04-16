package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func generateRandomUUID() (string, error) {
	var err error
	var randomUUID uuid.UUID
	for i := 0; i < 5; i++ {
		randomUUID, err = uuid.NewRandom()
		if err == nil {
			return randomUUID.String(), nil
		}
	}
	return "", err
}

// Deprecated: Use name value in place of id instead.
// getIDFromObjectMeta returns id from object meta
// id format: namespace/name
func getIDFromObjectMeta(meta metav1.ObjectMeta) string {
	return meta.Namespace + "/" + meta.Name
}

// Deprecated: Use getNameFromId() instead.
// getNameAndNamespaceFromId returns name and namespace from id of the
// resource or data source stored in a state
// id format: namespace/name
func getNameAndNamespaceFromId(id string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid id %s", id)
	}
	return parts[1], parts[0], nil
}

// getNameFromId returns name from id of the resource or data source stored in a state
// id format: namespace/name, this function ignores namespace.
func getNameFromId(id string) (string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		return id, nil
	}
	return parts[1], nil
}

// getNameFromNameOrId returns name from resource name or resource id
// We are phasing out the use of resource id and using name instead.
// For backward compatibility, we are using resource id if name is not provided.
func getNameFromNameOrId(name, id string) (string, error) {
	if name != "" {
		return name, nil
	}
	return getNameFromId(id)
}

func getNamespaceFromEnv() (string, error) {
	// TODO: Find a better way to get namespace from provider shared state/config
	namespace := os.Getenv("RXTSPOT_ORG_NS")
	if namespace == "" {
		return "", errors.New("RXTSPOT_ORG_NS is not set")
	}
	return namespace, nil
}

// FindNamespaceForOrganization returns namespace for organization
// ngpc API is used to find namespace
func FindNamespaceForOrganization(ctx context.Context, client ngpc.Client, orgName string) (string, error) {
	org, err := client.Organizer().LookupOrganizationByName(ctx, orgName)
	if err != nil {
		return "", err
	}
	if org == nil {
		return "", fmt.Errorf("organization %s not found", orgName)
	}
	return findNamespaceFromID(*org.ID), nil
}

func findNamespaceFromID(orgID string) string {
	return strings.ReplaceAll(strings.ToLower(orgID), "_", "-")
}

// readFileUpToNBytes reads file up to n bytes to prevent reading large files
func readFileUpToNBytes(filename string, n int64) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buf := make([]byte, n)
	_, err = io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	buf = bytes.Trim(buf, "\x00")

	return strings.TrimSpace(string(buf)), nil
}

// FindOrgName returns organization name from organization id
func FindOrgName(ctx context.Context, client ngpc.Client, userJWT string, orgID string) (string, error) {
	orgList, err := client.Organizer().ListOrganizationsForUser(ctx, userJWT)
	if err != nil {
		return "", err
	}
	if orgList == nil {
		return "", fmt.Errorf("organization list %s not found", orgID)
	}
	for _, org := range orgList.Organizations {
		if *org.ID == orgID {
			return org.GetDisplayName(), nil
		}
	}
	return "", fmt.Errorf("organization %s not found", orgID)
}
