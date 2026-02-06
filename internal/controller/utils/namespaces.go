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

package utils

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetOffloadedNamespaces returns the list of namespaces offloaded to the current provider.
// If the current cluster is a consumer, it returns an empty list.
func GetOffloadedNamespaces(ctx context.Context, cl client.Client, clusterID string) ([]corev1.Namespace, error) {
	selector := labels.NewSelector()
	reqEqual, _ := labels.NewRequirement("liqo.io/remote-cluster-id", selection.Equals, []string{clusterID})
	reqNotExist, _ := labels.NewRequirement("liqo.io/tenant-namespace", selection.DoesNotExist, nil)
	selector = selector.Add(*reqEqual, *reqNotExist)

	namespaceList := &corev1.NamespaceList{}
	if err := cl.List(ctx, namespaceList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return nil, err
	}

	return namespaceList.Items, nil
}
