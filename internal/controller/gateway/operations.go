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

package gateway

import (
	"context"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	connectivityv1 "github.com/riccardotornesello/liqo-connectivity-engine/api/v1"
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReconcileGatewayFirewallConfiguration ensures that the FirewallConfiguration
// resource for the gateway connectivity rules exists and is up to date.
// It creates or updates the resource as needed based on the provided
// PeeringConnectivity configuration.
func ReconcileGatewayFirewallConfiguration(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	cfg *connectivityv1.PeeringConnectivity,
	clusterID string,
) (controllerutil.OperationResult, error) {
	gatewayFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ForgeGatewayResourceName(clusterID),
			Namespace: utils.GetClusterNamespace(clusterID),
		},
	}

	return controllerutil.CreateOrUpdate(ctx, c, &gatewayFwcfg, func() error {
		// Set labels that identify this FirewallConfiguration as a gateway-level
		// connectivity configuration targeting all nodes.
		gatewayFwcfg.SetLabels(ForgeGatewayLabels(clusterID))

		// Generate the FirewallConfiguration spec based on the PeeringConnectivity rules.
		spec, err := ForgeGatewaySpec(ctx, c, cfg, clusterID)
		if err != nil {
			return err
		}
		gatewayFwcfg.Spec = *spec

		// Set owner reference so the FirewallConfiguration is deleted when the
		// PeeringConnectivity is deleted.
		return controllerutil.SetOwnerReference(cfg, &gatewayFwcfg, scheme)
	})
}

// EnsureGatewayFirewallConfigurationDeleted deletes the gateway-level FirewallConfiguration
// resource associated with the given cluster ID, if it exists.
func EnsureGatewayFirewallConfigurationDeleted(
	ctx context.Context,
	c client.Client,
	clusterID string,
) error {
	gatewayFwcfg := networkingv1beta1.FirewallConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ForgeGatewayResourceName(clusterID),
			Namespace: utils.GetClusterNamespace(clusterID),
		},
	}

	err := c.Delete(ctx, &gatewayFwcfg)
	if client.IgnoreNotFound(err) != nil {
		return err
	}
	return nil
}
