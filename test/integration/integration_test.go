/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"time"

	networkingv1beta1 "github.com/liqotech/liqo/apis/networking/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	securityv1 "github.com/riccardotornesello/liqo-security-manager/api/v1"
)

var _ = Describe("PeeringConnectivity Integration", func() {
	const (
		timeout  = time.Second * 30
		interval = time.Millisecond * 250
	)

	Context("When creating a PeeringConnectivity resource", func() {
		const (
			peeringName = "test-peering"
			clusterID   = "test-cluster"
		)

		var (
			namespace      string
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			namespace = "liqo-tenant-" + clusterID
			namespacedName = types.NamespacedName{
				Name:      peeringName,
				Namespace: namespace,
			}

			// Create the namespace
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil && !errors.IsAlreadyExists(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		})

		AfterEach(func() {
			// Cleanup
			peering := &securityv1.PeeringConnectivity{}
			err := k8sClient.Get(ctx, namespacedName, peering)
			if err == nil {
				Expect(k8sClient.Delete(ctx, peering)).To(Succeed())
				Eventually(func() bool {
					err := k8sClient.Get(ctx, namespacedName, peering)
					return errors.IsNotFound(err)
				}, timeout, interval).Should(BeTrue())
			}
		})

		It("should create FirewallConfiguration with allow rules", func() {
			By("Creating a PeeringConnectivity with allow rules")
			peering := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      peeringName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action:      securityv1.ActionAllow,
							Source:      &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
							Destination: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupLocalCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, peering)).To(Succeed())

			By("Verifying the FirewallConfiguration is created")
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}, timeout, interval).Should(Succeed())

			By("Verifying the FirewallConfiguration has correct structure")
			Expect(fwcfg.Spec.Table.Name).NotTo(BeNil())
			Expect(*fwcfg.Spec.Table.Name).To(Equal("cluster-security"))
			Expect(fwcfg.Spec.Table.Chains).To(HaveLen(1))
			Expect(fwcfg.Spec.Table.Chains[0].Rules.FilterRules).To(HaveLen(2)) // established + our rule

			By("Verifying the status is updated")
			Eventually(func() metav1.ConditionStatus {
				err := k8sClient.Get(ctx, namespacedName, peering)
				if err != nil {
					return metav1.ConditionUnknown
				}
				for _, cond := range peering.Status.Conditions {
					if cond.Type == "Ready" {
						return cond.Status
					}
				}
				return metav1.ConditionUnknown
			}, timeout, interval).Should(Equal(metav1.ConditionTrue))
		})

		It("should create FirewallConfiguration with deny rules", func() {
			By("Creating a PeeringConnectivity with deny rules")
			peering := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      peeringName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action: securityv1.ActionDeny,
							Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, peering)).To(Succeed())

			By("Verifying the FirewallConfiguration is created")
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}, timeout, interval).Should(Succeed())

			By("Verifying the FirewallConfiguration has deny action")
			Expect(fwcfg.Spec.Table.Chains[0].Rules.FilterRules).To(HaveLen(2)) // established + deny rule
		})

		It("should update FirewallConfiguration when PeeringConnectivity is updated", func() {
			By("Creating initial PeeringConnectivity")
			peering := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      peeringName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action: securityv1.ActionAllow,
							Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, peering)).To(Succeed())

			By("Waiting for initial FirewallConfiguration")
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}, timeout, interval).Should(Succeed())

			initialRuleCount := len(fwcfg.Spec.Table.Chains[0].Rules.FilterRules)

			By("Updating the PeeringConnectivity")
			Eventually(func() error {
				err := k8sClient.Get(ctx, namespacedName, peering)
				if err != nil {
					return err
				}
				peering.Spec.Rules = append(peering.Spec.Rules, securityv1.Rule{
					Action: securityv1.ActionDeny,
					Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupLocalCluster)},
				})
				return k8sClient.Update(ctx, peering)
			}, timeout, interval).Should(Succeed())

			By("Verifying the FirewallConfiguration is updated")
			Eventually(func() int {
				err := k8sClient.Get(ctx, fwcfgName, fwcfg)
				if err != nil {
					return 0
				}
				return len(fwcfg.Spec.Table.Chains[0].Rules.FilterRules)
			}, timeout, interval).Should(BeNumerically(">", initialRuleCount))
		})

		It("should handle multiple rules", func() {
			By("Creating a PeeringConnectivity with multiple rules")
			peering := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      peeringName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action:      securityv1.ActionAllow,
							Source:      &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
							Destination: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupLocalCluster)},
						},
						{
							Action:      securityv1.ActionAllow,
							Source:      &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupLocalCluster)},
							Destination: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
						{
							Action: securityv1.ActionDeny,
							Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, peering)).To(Succeed())

			By("Verifying the FirewallConfiguration is created with all rules")
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}, timeout, interval).Should(Succeed())

			// Should have 1 established rule + 3 custom rules = 4 total
			Expect(fwcfg.Spec.Table.Chains[0].Rules.FilterRules).To(HaveLen(4))
		})

		It("should delete FirewallConfiguration when PeeringConnectivity is deleted", func() {
			By("Creating a PeeringConnectivity")
			peering := &securityv1.PeeringConnectivity{
				ObjectMeta: metav1.ObjectMeta{
					Name:      peeringName,
					Namespace: namespace,
				},
				Spec: securityv1.PeeringConnectivitySpec{
					Rules: []securityv1.Rule{
						{
							Action: securityv1.ActionAllow,
							Source: &securityv1.Party{Group: ptr.To(securityv1.ResourceGroupRemoteCluster)},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, peering)).To(Succeed())

			By("Verifying the FirewallConfiguration is created")
			fwcfgName := types.NamespacedName{
				Name:      clusterID + "-security-fabric",
				Namespace: namespace,
			}
			fwcfg := &networkingv1beta1.FirewallConfiguration{}
			Eventually(func() error {
				return k8sClient.Get(ctx, fwcfgName, fwcfg)
			}, timeout, interval).Should(Succeed())

			By("Deleting the PeeringConnectivity")
			Expect(k8sClient.Delete(ctx, peering)).To(Succeed())

			By("Verifying the FirewallConfiguration is deleted")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, fwcfgName, fwcfg)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
