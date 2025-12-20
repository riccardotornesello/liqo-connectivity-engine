package controller

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

// Remove the tenantnamespace.NamePrefix + "-" from the namespace to get the cluster ID
func extractClusterID(namespace string) (string, error) {
	const prefix = tenantnamespace.NamePrefix + "-"

	if len(namespace) <= len(prefix) || namespace[:len(prefix)] != prefix {
		return "", fmt.Errorf("namespace %q does not have the expected prefix %q", namespace, prefix)
	}
	return namespace[len(prefix):], nil
}

func forgeGatewayResourceName(clusterID string) string {
	return fmt.Sprintf("%s-%s", clusterID, gatewayResourceNameSuffix)
}

func forgeGatewayLabels(clusterID string) map[string]string {
	// TODO: liqo managed?
	// TODO: category security?

	return map[string]string{
		firewall.FirewallCategoryTargetKey: gateway.FirewallCategoryGwTargetValue,
		firewall.FirewallUniqueTargetKey:   clusterID,
	}
}

func forgeGatewaySpec(cfg *securityv1.PeeringSecurity) (*networkingv1beta1.FirewallConfigurationSpec, error) {
	var policy networkingv1beta1firewall.ChainPolicy

	switch cfg.Spec.TunnelPolicy {
	case securityv1.TunnelPolicyAllow:
		policy = networkingv1beta1firewall.ChainPolicyAccept
	case securityv1.TunnelPolicyDeny:
		policy = networkingv1beta1firewall.ChainPolicyDrop
	default:
		return nil, fmt.Errorf("unknown tunnel policy %q", cfg.Spec.TunnelPolicy)
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
				// TODO: Type:     ptr.To(networkingv1beta1firewall.ChainTypeFilter),
				// TODO: Rules:    networkingv1beta1firewall.RulesSet{},
			}},
		},
	}, nil
}
