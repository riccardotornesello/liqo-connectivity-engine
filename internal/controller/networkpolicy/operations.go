// Copyright 2019-2026 The Liqo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networkpolicy

import (
	"context"

	"github.com/liqotech/liqo/pkg/consts"
	connectivityv1 "github.com/riccardotornesello/liqo-connectivity-engine/api/v1"
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller/utils"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	networkPolicyName = "liqo-connectivity-network-policy"
)

func ReconcileNetworkPolicies(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	cfg *connectivityv1.PeeringConnectivity,
	clusterID string,
) error {
	namespaces, err := utils.GetOffloadedNamespaces(ctx, c, clusterID)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		if _, err := reconcileNetworkPolicyInNamespace(ctx, c, scheme, cfg, clusterID, ns.Name); err != nil {
			return err
		}
	}

	return nil
}

// reconcileNetworkPolicyInNamespace ensures that the NetworkPolicy exists in the given namespace
// with the correct specification based on the PeeringConnectivity configuration.
func reconcileNetworkPolicyInNamespace(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	cfg *connectivityv1.PeeringConnectivity,
	clusterID string,
	namespaceName string,
) (controllerutil.OperationResult, error) {
	networkPolicy := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyName,
			Namespace: namespaceName,
		},
	}

	return controllerutil.CreateOrUpdate(ctx, c, &networkPolicy, func() error {
		// Set labels
		networkPolicy.SetLabels(map[string]string{
			consts.RemoteClusterID: clusterID,
		})

		// Generate the NetworkPolicy spec based on the PeeringConnectivity rules.
		spec, err := ForgeProviderNetworkPolicySpec(ctx, c, cfg, clusterID)
		if err != nil {
			return err
		}
		networkPolicy.Spec = *spec

		return nil
	})
}

// EnsureNetworkPoliciesDeleted deletes the NetworkPolicy resources
// associated with the given cluster ID, if they exist.
func EnsureNetworkPoliciesDeleted(
	ctx context.Context,
	c client.Client,
	clusterID string,
) error {
	namespaces, err := utils.GetOffloadedNamespaces(ctx, c, clusterID)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		if err := deleteNetworkPolicyInNamespace(ctx, c, ns.Name); err != nil {
			return err
		}
	}

	return nil
}

func deleteNetworkPolicyInNamespace(
	ctx context.Context,
	c client.Client,
	namespaceName string,
) error {
	networkPolicy := networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      networkPolicyName,
			Namespace: namespaceName,
		},
	}

	err := c.Delete(ctx, &networkPolicy)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}
