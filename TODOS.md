# Known TODOs and Future Improvements

This document tracks known TODOs in the codebase and plans for future improvements.

## Current TODOs

### Code TODOs

1. **Pod Label Optimization** (`internal/controller/utils/pods.go`)
   - **TODO**: Optimize by adding labels in Liqo when offloading pods
   - **Impact**: Currently we filter pods by checking all pods with certain labels and then filter by node name
   - **Improvement**: If Liqo adds a label identifying the provider cluster ID, we can filter more efficiently
   - **Priority**: Low - Current implementation works, this is just an optimization
   - **Depends on**: Liqo upstream changes

2. **Certificate Manager Integration** (`cmd/main.go`)
   - **TODO**: Enable cert-manager integration for metrics endpoint
   - **Impact**: Currently using self-signed certificates for metrics in development
   - **Improvement**: Use cert-manager for production-grade certificate management
   - **Priority**: Medium - Important for production deployments
   - **Action**: Uncomment the cert-manager configuration in `config/default/kustomization.yaml` and `config/prometheus/kustomization.yaml`

3. **Test Customization** (test files)
   - **TODO**: Add more specific test scenarios
   - **Impact**: Current tests are basic scaffolding from Kubebuilder
   - **Improvement**: Add comprehensive unit and e2e tests for all scenarios
   - **Priority**: High - Critical for production readiness
   - **See**: Testing section below

## Roadmap Items

### Short-term (Next Release)

- [ ] **Webhook Validation**: Add admission webhooks to validate PeeringConnectivity resources
  - Validate resource group combinations
  - Prevent conflicting rules
  - Ensure proper namespace format

- [ ] **Enhanced Status**: Add more detailed status information
  - Number of rules applied
  - Number of pods affected
  - Last sync timestamp
  - Health of FirewallConfiguration

- [ ] **Metrics**: Add Prometheus metrics
  - Number of PeeringConnectivity resources
  - Number of rules per resource
  - Reconciliation duration
  - Error rates

### Medium-term (Future Releases)

- [ ] **IPv6 Support**: Add support for IPv6 CIDRs and addresses
  - Update CIDR retrieval functions
  - Update firewall rule generation
  - Test with dual-stack clusters

- [ ] **Port and Protocol Filtering**: Add support for port and protocol-level rules
  - Extend API to support port ranges
  - Add protocol specifications (TCP, UDP, ICMP)
  - Update FirewallConfiguration generation

- [ ] **Named Resource Groups**: Allow users to define custom resource groups
  - Custom pod selectors
  - Namespace-based groups
  - Label-based groups

- [ ] **Multiple PeeringConnectivity per Cluster**: Support multiple policies per cluster
  - Currently assumes one policy per tenant namespace
  - Allow for use-case specific policies
  - Handle policy conflicts and precedence

### Long-term (Research & Development)

- [ ] **Network Policy Integration**: Integrate with Kubernetes NetworkPolicies
  - Complement NetworkPolicies with cross-cluster rules
  - Automatic conversion between formats
  - Unified policy model

- [ ] **Policy Simulation**: Add dry-run mode to test policies before applying
  - Simulate traffic flows
  - Identify blocked connections
  - Suggest policy improvements

- [ ] **Observability Dashboard**: Create a UI for policy management
  - Visualize traffic flows
  - Monitor policy effectiveness
  - Simplified policy creation

- [ ] **Policy Templates**: Provide pre-built policy templates
  - Common use cases
  - Security best practices
  - Quick start configurations

## Testing Strategy

### Unit Tests Needed

- [ ] API validation tests
- [ ] Resource group function tests
- [ ] Firewall rule generation tests
- [ ] Pod filtering tests
- [ ] CIDR retrieval tests
- [ ] Controller reconciliation logic tests

### Integration Tests Needed

- [ ] End-to-end policy application
- [ ] Multi-cluster scenarios
- [ ] Pod lifecycle events
- [ ] Network changes
- [ ] Namespace offloading changes

### E2E Tests Needed

- [ ] Consumer cluster scenarios
- [ ] Provider cluster scenarios
- [ ] Multi-tenant provider
- [ ] Policy updates and rollbacks
- [ ] Failure and recovery scenarios

## Dependencies

### Current Dependencies

All dependencies are up-to-date as of the last review. Key dependencies:

- **Liqo**: v1.0.3 (replaced with local development version)
- **Kubernetes**: v0.34.1
- **controller-runtime**: v0.22.4

### Dependency Notes

1. **Liqo Local Replace**: The go.mod includes `replace github.com/liqotech/liqo => ../liqo`
   - This is for development purposes
   - Before release, ensure this points to a stable Liqo version
   - Consider vendoring if needed

2. **Go Version**: Currently using Go 1.25.5
   - Keep in sync with Liqo requirements
   - Update as new Go versions are released

## Contributing to TODOs

If you'd like to work on any of these TODOs:

1. Check if there's already an issue for it
2. If not, create an issue describing your approach
3. Reference the issue in your pull request
4. Update this document when completing a TODO

## Reporting New TODOs

When adding a new TODO to the code:

1. Add a descriptive comment: `// TODO: Description of what needs to be done`
2. Create an issue to track it
3. Add it to this document if it's significant
4. Consider the priority and impact

## Documentation TODOs

- [ ] Add video tutorials
- [ ] Add architecture diagrams
- [ ] Add performance benchmarks
- [ ] Add security audit results
- [ ] Add migration guides
- [ ] Add troubleshooting playbooks
