// Package utils provides utility functions for creating firewall sets in Liqo.
// A set is a collection of IP addresses that can be referenced in firewall rules.
package utils

import (
	networkingv1beta1firewall "github.com/liqotech/liqo/apis/networking/v1beta1/firewall"
	corev1 "k8s.io/api/core/v1"
)

// ForgePodIpsSet creates a firewall Set containing the IP addresses of the given pods.
// This set can be referenced in firewall rules to match traffic to/from these pods.
// Pods without an IP address are excluded from the set.
func ForgePodIpsSet(setName string, pods []corev1.Pod) networkingv1beta1firewall.Set {
	setElements := make([]networkingv1beta1firewall.SetElement, 0)
	for _, pod := range pods {
		podIp := pod.Status.PodIP
		if podIp == "" {
			// Skip pods that don't have an IP address yet.
			continue
		}
		setElements = append(setElements, networkingv1beta1firewall.SetElement{
			Key: podIp,
		})
	}

	return networkingv1beta1firewall.Set{
		Name:     setName,
		KeyType:  networkingv1beta1firewall.SetDataTypeIPAddr,
		Elements: setElements,
	}
}
