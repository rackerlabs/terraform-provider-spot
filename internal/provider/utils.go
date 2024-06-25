package provider

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/google/uuid"
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
	// TODO: Find a better way to get value of the namespace from provider.go
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

func StrSliceContains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func matchesExpression[T int | float64](expression string, value T) bool {
	expression = strings.TrimSpace(expression)

	if operandStr, found := strings.CutPrefix(expression, "=="); found {
		operand, err := strconv.ParseFloat(strings.TrimSpace(operandStr), 64)
		if err != nil {
			return false
		}
		return float64(value) == operand
	}

	if operandStr, found := strings.CutPrefix(expression, "<="); found {
		operand, err := strconv.ParseFloat(strings.TrimSpace(operandStr), 64)
		if err != nil {
			return false
		}
		return float64(value) <= operand
	}

	if operandStr, found := strings.CutPrefix(expression, ">="); found {
		operand, err := strconv.ParseFloat(strings.TrimSpace(operandStr), 64)
		if err != nil {
			return false
		}
		return float64(value) >= operand
	}

	if operandStr, found := strings.CutPrefix(expression, "<"); found {
		operand, err := strconv.ParseFloat(strings.TrimSpace(operandStr), 64)
		if err != nil {
			return false
		}
		return float64(value) < operand
	}

	if operandStr, found := strings.CutPrefix(expression, ">"); found {
		operand, err := strconv.ParseFloat(strings.TrimSpace(operandStr), 64)
		if err != nil {
			return false
		}
		return float64(value) > operand
	}

	if operandStr, found := strings.CutPrefix(expression, "!="); found {
		operand, err := strconv.ParseFloat(strings.TrimSpace(operandStr), 64)
		if err != nil {
			return false
		}
		return float64(value) != operand
	}
	operand, err := strconv.ParseFloat(expression, 64)
	if err != nil {
		return false
	}
	return float64(value) == operand
}
