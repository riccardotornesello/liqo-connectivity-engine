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
	"fmt"

	connectivityv1 "github.com/riccardotornesello/liqo-connectivity-engine/api/v1"
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/resourcegroups"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ForgeProviderNetworkPolicySpec(
	ctx context.Context,
	cl client.Client,
	cfg *connectivityv1.PeeringConnectivity,
	clusterID string,
) (*networkingv1.NetworkPolicySpec, error) {
	spec := networkingv1.NetworkPolicySpec{
		Ingress:     []networkingv1.NetworkPolicyIngressRule{},
		Egress:      []networkingv1.NetworkPolicyEgressRule{},
		PolicyTypes: []networkingv1.PolicyType{networkingv1.PolicyTypeIngress, networkingv1.PolicyTypeEgress},
	}

	// Add rules based on the PeeringConnectivity configuration.
	for _, rule := range cfg.Spec.Rules {
		if rule.Source != nil && rule.Source.Group != nil && *rule.Source.Group == connectivityv1.ResourceGroupOffloaded {
			to, toPorts, err := ForgeNetworkPolicyPeer(ctx, cl, clusterID, rule.Destination)
			if err != nil {
				return nil, fmt.Errorf("failed to forge network policy peer for rule destination: %w", err)
			}
			spec.Egress = append(spec.Egress, networkingv1.NetworkPolicyEgressRule{To: to, Ports: toPorts})
		}

		if rule.Destination != nil && rule.Destination.Group != nil && *rule.Destination.Group == connectivityv1.ResourceGroupOffloaded {
			from, fromPorts, err := ForgeNetworkPolicyPeer(ctx, cl, clusterID, rule.Source)
			if err != nil {
				return nil, fmt.Errorf("failed to forge network policy peer for rule source: %w", err)
			}
			spec.Ingress = append(spec.Ingress, networkingv1.NetworkPolicyIngressRule{From: from, Ports: fromPorts})
		}
	}

	return &spec, nil
}

func ForgeNetworkPolicyPeer(ctx context.Context, cl client.Client, clusterID string, peer *connectivityv1.Party) ([]networkingv1.NetworkPolicyPeer, []networkingv1.NetworkPolicyPort, error) {
	if peer == nil {
		return nil, nil, fmt.Errorf("party is nil")
	}

	if peer.Namespace != nil {
		return []networkingv1.NetworkPolicyPeer{{
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"kubernetes.io/metadata.name": *peer.Namespace,
				},
			},
		}}, nil, nil
	}

	if peer.Group != nil {
		return resourcegroups.ResourceGroupFuncts[*peer.Group].MakeNetworkPolicyRule(ctx, cl, clusterID)
	}

	return nil, nil, fmt.Errorf("unsupported party configuration: %+v", peer)
}
