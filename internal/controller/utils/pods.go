package utils

import (
	"context"

	offloadingv1beta1 "github.com/liqotech/liqo/apis/offloading/v1beta1"
	"github.com/liqotech/liqo/pkg/consts"
	"github.com/liqotech/liqo/pkg/virtualKubelet/forge"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Function for the Consumer. Returns the list of Pod IPs on the remote Provider cluster.
func GetPodsOffloadedToProvider(ctx context.Context, cl client.Client, providerClusterID string) ([]corev1.Pod, error) {
	// TODO: optimize by adding labels in liqo when offloading pods

	// Get all the pods offloaded to the provider cluster.
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.MatchingLabels{
		consts.LocalPodLabelKey: consts.LocalPodLabelValue,
	}); err != nil {
		return nil, err
	}

	// Filter the pods owned by the provider cluster.
	pods := make([]corev1.Pod, 0)
	for _, pod := range podList.Items {
		if pod.Spec.NodeName == providerClusterID {
			pods = append(pods, pod)
		}
	}

	return pods, nil
}

// Function for the Provider. Returns the list of Pods owned by the Consumer.
func GetPodsFromConsumer(ctx context.Context, cl client.Client, consumerClusterID string) ([]corev1.Pod, error) {
	// Get the pods coming from the remote cluster.
	podList := &corev1.PodList{}
	if err := cl.List(ctx, podList, client.MatchingLabels{
		forge.LiqoOriginClusterIDKey: consumerClusterID,
	}); err != nil {
		return nil, err
	}

	return podList.Items, nil
}

// Get the Pods in offloaded namespaces hosted on the consumer cluster.
func GetPodsInOffloadedNamespaces(ctx context.Context, cl client.Client) ([]corev1.Pod, error) {
	// Get all the namespaces offloaded.
	namespaceList := &offloadingv1beta1.NamespaceOffloadingList{}
	if err := cl.List(ctx, namespaceList); err != nil {
		return nil, err
	}

	// Get all the pods in the offloaded namespaces. Exclude local shadow pods.
	var pods []corev1.Pod

	// NotEquals includes also the case where the label is not present.
	labelsRequirement, err := labels.NewRequirement(consts.LocalPodLabelKey, selection.NotEquals, []string{consts.LocalPodLabelValue})
	if err != nil {
		return nil, err
	}

	for _, nso := range namespaceList.Items {
		podList := &corev1.PodList{}
		if err := cl.List(
			ctx,
			podList,
			client.InNamespace(nso.Namespace),
			client.MatchingLabelsSelector{Selector: labels.NewSelector().Add(*labelsRequirement)},
		); err != nil {
			return nil, err
		}
		pods = append(pods, podList.Items...)
	}

	return pods, nil
}
