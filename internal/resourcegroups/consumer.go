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

package resourcegroups

import (
	"context"

	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	"github.com/liqotech/liqo/pkg/consts"
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller/utils"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// slice-local: Matches local pods in namespaces that are configured for offloading.
// These are the actual pods running locally that could be offloaded.
// Uses a set because pod IPs are dynamically allocated.
var ResourceGroupSliceLocal = groupFuncts{
	MakeFirewallConfigurationSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
		// Get all pods in namespaces that are configured for offloading.
		pods, err := utils.GetPodsInOffloadedNamespaces(ctx, cl)
		if err != nil {
			return nil, err
		}

		// Create a firewall set containing the IPs of these pods.
		podIpsSet := utils.ForgePodIpsSet("vclocal", pods)
		return []networkingv1beta1firewall.Set{podIpsSet}, nil
	},
	MakeFirewallConfigurationRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
		return []networkingv1beta1firewall.Match{{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    "@vclocal",
				Position: position,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		}}, nil
	},
	// TODO: make networkpolicy
}

// slice-remote: Matches shadow pods on the consumer cluster that represent
// pods offloaded to a provider cluster.
// Uses a set because pod IPs are dynamically allocated.
var ResourceGroupSliceRemote = groupFuncts{
	MakeFirewallConfigurationSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
		// Get all shadow pods that represent offloaded pods on the specified provider cluster.
		pods, err := utils.GetPodsOffloadedToProvider(ctx, cl, clusterID)
		if err != nil {
			return nil, err
		}

		// Create a firewall set containing the IPs of these shadow pods.
		podIpsSet := utils.ForgePodIpsSet("vcremote", pods)
		return []networkingv1beta1firewall.Set{podIpsSet}, nil
	},
	MakeFirewallConfigurationRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
		return []networkingv1beta1firewall.Match{{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    "@vcremote",
				Position: position,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		}}, nil
	},
	MakeNetworkPolicyRule: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1.NetworkPolicyPeer, []networkingv1.NetworkPolicyPort, error) {
		return []networkingv1.NetworkPolicyPeer{{
			PodSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					consts.LocalPodLabelKey: consts.LocalPodLabelValue,
				},
			},
		}}, nil, nil
	},
}
