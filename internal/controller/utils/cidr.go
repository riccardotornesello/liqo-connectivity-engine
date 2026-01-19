// Package utils provides utility functions for working with CIDR blocks in Liqo clusters.
// It retrieves pod CIDR information from Liqo Network resources for both local and remote clusters.
package utils

import (
	"context"
	"fmt"

	ipamv1alpha1 "github.com/liqotech/liqo/apis/ipam/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// localPodCIDRNetworkName is the name of the Network resource that contains
	// the local cluster's pod CIDR information.
	localPodCIDRNetworkName = "pod-cidr"
	// localPodCIDRNetworkNamespace is the namespace where the local pod CIDR Network resource is stored.
	localPodCIDRNetworkNamespace = "liqo"
)

// GetCurrentClusterPodCIDR retrieves the pod CIDR for the current (local) cluster.
// It reads the Network resource in the liqo namespace to obtain the CIDR.
func GetCurrentClusterPodCIDR(ctx context.Context, cl client.Client) (string, error) {
	var network ipamv1alpha1.Network

	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: localPodCIDRNetworkNamespace,
		Name:      localPodCIDRNetworkName,
	}, &network); err != nil {
		return "", err
	}

	return string(network.Spec.CIDR), nil
}

// GetRemoteClusterPodCIDR retrieves the pod CIDR for a remote peered cluster.
// It reads the Network resource in the tenant namespace for the specified cluster ID.
func GetRemoteClusterPodCIDR(ctx context.Context, cl client.Client, clusterID string) (string, error) {
	var network ipamv1alpha1.Network

	if err := cl.Get(ctx, client.ObjectKey{
		Namespace: GetClusterNamespace(clusterID),
		Name:      fmt.Sprintf("%s-pod", clusterID),
	}, &network); err != nil {
		return "", err
	}

	return string(network.Status.CIDR), nil
}
