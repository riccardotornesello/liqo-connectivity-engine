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
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// internet: Matches traffic destined to all public IP ranges except those described in RFC1918.
var ResourceGroupInternet = groupFuncts{
	MakeFirewallConfigurationSets: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1beta1firewall.Set, error) {
		return []networkingv1beta1firewall.Set{{
			Name:    "privatesubnets",
			KeyType: networkingv1beta1firewall.SetDataTypeIPCIDR,
			Elements: []networkingv1beta1firewall.SetElement{
				{Key: "10.0.0.0/8"},
				{Key: "172.16.0.0/12"},
				{Key: "192.168.0.0/16"},
			},
		}}, nil
	},
	MakeFirewallConfigurationRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
		return []networkingv1beta1firewall.Match{{
			IP: &networkingv1beta1firewall.MatchIP{
				Value:    "@privatesubnets",
				Position: position,
			},
			Op: networkingv1beta1firewall.MatchOperationNeq,
		}}, nil
	},
	MakeNetworkPolicyRule: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1.NetworkPolicyPeer, []networkingv1.NetworkPolicyPort, error) {
		return []networkingv1.NetworkPolicyPeer{{
			IPBlock: &networkingv1.IPBlock{
				CIDR: "0.0.0.0/0",
				Except: []string{
					"10.0.0.0/8",
					"172.16.0.0/12",
					"192.168.0.0/16",
				},
			},
		}}, nil, nil
	},
}

// nameserver: Matches traffic destined to any nameserver (port 53).
var ResourceGroupNameserver = groupFuncts{
	MakeFirewallConfigurationSets: nil, // No sets needed for this group
	MakeFirewallConfigurationRule: func(ctx context.Context, cl client.Client, clusterID string, position networkingv1beta1firewall.MatchPosition) ([]networkingv1beta1firewall.Match, error) {
		return []networkingv1beta1firewall.Match{{
			Port: &networkingv1beta1firewall.MatchPort{
				Value:    "53",
				Position: position,
			},
			Op: networkingv1beta1firewall.MatchOperationEq,
		}}, nil
	},
	MakeNetworkPolicyRule: func(ctx context.Context, cl client.Client, clusterID string) ([]networkingv1.NetworkPolicyPeer, []networkingv1.NetworkPolicyPort, error) {
		return nil, []networkingv1.NetworkPolicyPort{{
			Port:     ptr.To(intstr.FromInt(53)),
			Protocol: ptr.To(corev1.ProtocolTCP),
		}, {
			Port:     ptr.To(intstr.FromInt(53)),
			Protocol: ptr.To(corev1.ProtocolUDP),
		}}, nil
	},
}
