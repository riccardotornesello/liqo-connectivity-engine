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

func mapTunnelPolicyToChainPolicy(policy securityv1.TunnelPolicy) (networkingv1beta1firewall.ChainPolicy, error) {
	switch policy {
	case securityv1.TunnelPolicyAllow:
		return networkingv1beta1firewall.ChainPolicyAccept, nil
	case securityv1.TunnelPolicyDeny:
		return networkingv1beta1firewall.ChainPolicyDrop, nil
	default:
		return "", fmt.Errorf("unknown tunnel policy %q", policy)
	}
}

func mapPolicyRuleGroup(group securityv1.ResourceGroup) (string, error) {
	switch group {
	case securityv1.ResourceGroupVcLocal:
		return fmt.Sprintf("@%s", gatewayVcLocalPodsSetName), nil
	case securityv1.ResourceGroupVcRemote:
		return fmt.Sprintf("@%s", gatewayVcRemotePodsSetName), nil
	// TODO: securityv1.ResourceGroupRemote
	// TODO: securityv1.ResourceGroupOffloaded
	default:
		return "", fmt.Errorf("unknown resource group %q", group)
	}
}

func forgeGatewaySpec(cfg *securityv1.PeeringSecurity) (*networkingv1beta1.FirewallConfigurationSpec, error) {
	var filterRules []networkingv1beta1firewall.FilterRule

	policy, err := mapTunnelPolicyToChainPolicy(cfg.Spec.TunnelPolicy)
	if err != nil {
		return nil, err
	}

	for _, rule := range cfg.Spec.Rules {
		var action networkingv1beta1firewall.FilterAction
		var match []networkingv1beta1firewall.Match

		switch rule.Action {
		case securityv1.RuleActionAllow:
			action = networkingv1beta1firewall.ActionAccept
		case securityv1.RuleActionDeny:
			action = networkingv1beta1firewall.ActionDrop
		default:
			return nil, fmt.Errorf("unknown rule action %q", rule.Action)
		}

		if rule.Src != nil {
			ip, err := mapPolicyRuleGroup(*rule.Src)
			if err != nil {
				return nil, err
			}

			match = append(match, networkingv1beta1firewall.Match{
				IP: &networkingv1beta1firewall.MatchIP{
					Position: networkingv1beta1firewall.MatchPositionSrc,
					Value:    ip,
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			})
		}

		if rule.Dst != nil {
			ip, err := mapPolicyRuleGroup(*rule.Dst)
			if err != nil {
				return nil, err
			}

			match = append(match, networkingv1beta1firewall.Match{
				IP: &networkingv1beta1firewall.MatchIP{
					Position: networkingv1beta1firewall.MatchPositionDst,
					Value:    ip,
				},
				Op: networkingv1beta1firewall.MatchOperationEq,
			})
		}

		filterRules = append(filterRules,
			networkingv1beta1firewall.FilterRule{
				Action: action,
				Match:  match,
			})
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
