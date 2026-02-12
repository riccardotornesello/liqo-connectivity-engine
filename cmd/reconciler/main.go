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

// Package main executes a standalone run of the PeeringConnectivityReconciler.
// It is intended for testing and debugging purposes, allowing developers to run
// the reconciler logic in isolation without deploying the full controller manager.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller"
	"github.com/riccardotornesello/liqo-connectivity-engine/internal/controller/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func main() {
	var clusterID string

	flag.StringVar(&clusterID, "cluster-id", "", "The ID of the cluster to test the controller with.")
	flag.Parse()

	if clusterID == "" {
		fmt.Println("Error: cluster-id flag is required")
		os.Exit(1)
	}

	opts := zap.Options{Development: true}
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg := config.GetConfigOrDie()

	scheme := runtime.NewScheme()
	utils.RegisterScheme(scheme)

	cl, _ := client.New(cfg, client.Options{Scheme: scheme})

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	recorder := eventBroadcaster.NewRecorder(scheme, v1.EventSource{Component: "my-manual-controller"})

	reconciler := &controller.PeeringConnectivityReconciler{
		Client:   cl,
		Scheme:   scheme,
		Recorder: recorder,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      clusterID,
			Namespace: fmt.Sprintf("liqo-tenant-%s", clusterID),
		},
	}

	res, err := reconciler.Reconcile(context.Background(), req)
	fmt.Printf("Result: %+v, Error: %v\n", res, err)

	// Exit with an error code if reconciliation failed.
	if err != nil {
		os.Exit(1)
	}
}
