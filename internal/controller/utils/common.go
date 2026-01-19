// Package utils provides utility functions for the PeeringConnectivity controller.
// It includes functions for extracting cluster IDs, retrieving CIDR information,
// managing resource groups, and working with pod collections.
package utils

import (
	"fmt"

	tenantnamespace "github.com/liqotech/liqo/pkg/tenantNamespace"
)

// GetClusterNamespace returns the Liqo tenant namespace for a given cluster ID.
// Liqo uses namespaces with the format "liqo-tenant-<cluster-id>" to isolate
// resources for each peered cluster.
func GetClusterNamespace(clusterID string) string {
	return fmt.Sprintf("%s-%s", tenantnamespace.NamePrefix, clusterID)
}

// ExtractClusterID extracts the cluster ID from a Liqo tenant namespace name.
// It removes the "liqo-tenant-" prefix to obtain the cluster ID.
// Returns an error if the namespace doesn't follow the expected format.
func ExtractClusterID(namespace string) (string, error) {
	const prefix = tenantnamespace.NamePrefix + "-"

	if len(namespace) <= len(prefix) || namespace[:len(prefix)] != prefix {
		return "", fmt.Errorf("namespace %q does not have the expected prefix %q", namespace, prefix)
	}
	return namespace[len(prefix):], nil
}
