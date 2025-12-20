package controller

const (
	// gatewayTableName is the name of the firewall table used by the gateway FirewallConfiguration.
	gatewayTableName = "cluster-security"

	// gatewayResourceNameSuffix is the suffix appended to the cluster ID to form the gateway FirewallConfiguration name.
	gatewayResourceNameSuffix = "security-gateway"

	// gatewayChainName is the name of the firewall chain used by the gateway FirewallConfiguration.
	gatewayChainName = "cluster-security-filter"

	// gatewayChainPriority is the priority of the gateway firewall chain.
	gatewayChainPriority = 200

	// gatewayVcLocalPodsSetName is the name of the IPSet containing the local virtual cluster pods' IPs.
	gatewayVcLocalPodsSetName = "vc-local-pods"

	// gatewayVcRemotePodsSetName is the name of the IPSet containing the remote virtual cluster pods' IPs.
	gatewayVcRemotePodsSetName = "vc-remote-pods"
)
