package utils

import (
	"fmt"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	"github.com/liqotech/liqo/pkg/firewall"
	"github.com/liqotech/liqo/pkg/gateway"
	tenantnamespace "github.com/liqotech/liqo/pkg/tenantNamespace"
	"k8s.io/utils/ptr"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

const (
	// gatewayResourceNameSuffix is the suffix appended to the cluster ID to form the gateway FirewallConfiguration name.
	gatewayResourceNameSuffix = "security-gateway"

	// gatewayTableName is the name of the firewall table used by the gateway FirewallConfiguration.
	gatewayTableName = "cluster-security"

	// gatewayChainName is the name of the firewall chain used by the gateway FirewallConfiguration.
	gatewayChainName = "cluster-security-filter"

	// gatewayChainPriority is the priority of the gateway firewall chain.
	gatewayChainPriority = 200
)

// Generate the name of the Gateway FirewallConfiguration resource for the given cluster ID.
func ForgeGatewayResourceName(clusterID string) string {
	return fmt.Sprintf("%s-%s", clusterID, gatewayResourceNameSuffix)
}

func ForgeGatewayLabels(clusterID string) map[string]string {
	// TODO: liqo managed?
	// TODO: category security?

	return map[string]string{
		firewall.FirewallCategoryTargetKey: gateway.FirewallCategoryGwTargetValue,
		firewall.FirewallUniqueTargetKey:   clusterID,
	}
}

func ForgeGatewaySpec(cfg *securityv1.PeeringSecurity) (*networkingv1beta1.FirewallConfigurationSpec, error) {
	var policy networkingv1beta1firewall.ChainPolicy
	filterRules := []networkingv1beta1firewall.FilterRule{
		{
			Name:   ptr.To("allow-established-related"),
			Action: networkingv1beta1firewall.ActionAccept,
			Match: []networkingv1beta1firewall.Match{{
				CtState: &networkingv1beta1firewall.MatchCtState{
					Value: []networkingv1beta1firewall.CtStateValue{
						networkingv1beta1firewall.CtStateEstablished,
						networkingv1beta1firewall.CtStateRelated,
					},
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			}},
		},
		{
			Name:   ptr.To("only-from-tunnel"),
			Action: networkingv1beta1firewall.ActionAccept,
			Match: []networkingv1beta1firewall.Match{{
				Dev: &networkingv1beta1firewall.MatchDev{
					Position: networkingv1beta1firewall.MatchDevPositionIn,
					Value:    "liqo-tunnel",
				},
				Op: networkingv1beta1firewall.MatchOperationNeq,
			}},
		},
		{
			Name:   ptr.To("allow-eth"),
			Action: networkingv1beta1firewall.ActionAccept,
			Match: []networkingv1beta1firewall.Match{{
				Dev: &networkingv1beta1firewall.MatchDev{
					Position: networkingv1beta1firewall.MatchDevPositionOut,
					Value:    "eth0",
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			}},
		},
	}

	if cfg.Spec.BlockTunnelTraffic {
		policy = networkingv1beta1firewall.ChainPolicyDrop
	} else {
		policy = networkingv1beta1firewall.ChainPolicyAccept
	}

	return &networkingv1beta1.FirewallConfigurationSpec{
		Table: networkingv1beta1firewall.Table{
			Name:   ptr.To(gatewayTableName),
			Family: ptr.To(networkingv1beta1firewall.TableFamilyIPv4),
			Chains: []networkingv1beta1firewall.Chain{{
				Name:     ptr.To(gatewayChainName),
				Hook:     ptr.To(networkingv1beta1firewall.ChainHookPostrouting),
				Policy:   ptr.To(policy),
				Priority: ptr.To[networkingv1beta1firewall.ChainPriority](gatewayChainPriority),
				Type:     ptr.To(networkingv1beta1firewall.ChainTypeFilter),
				Rules: networkingv1beta1firewall.RulesSet{
					FilterRules: filterRules,
				},
			}},
		},
	}, nil
}

// ExtractClusterID extracts the cluster ID from the given namespace.
func ExtractClusterID(namespace string) (string, error) {
	// Remove the tenantnamespace.NamePrefix + "-" from the namespace to get the cluster ID
	const prefix = tenantnamespace.NamePrefix + "-"

	if len(namespace) <= len(prefix) || namespace[:len(prefix)] != prefix {
		return "", fmt.Errorf("namespace %q does not have the expected prefix %q", namespace, prefix)
	}
	return namespace[len(prefix):], nil
}
