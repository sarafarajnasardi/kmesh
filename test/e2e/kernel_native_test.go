//go:build integ
// +build integ

/*
 * Copyright The Kmesh Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kmesh

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"istio.io/istio/pkg/test/framework"
	"istio.io/istio/pkg/test/framework/components/echo"
	"istio.io/istio/pkg/test/framework/components/echo/check"
	"istio.io/istio/pkg/test/framework/components/echo/common/ports"
	"istio.io/istio/pkg/test/framework/components/echo/deployment"
	"istio.io/istio/pkg/test/framework/components/echo/match"
	"istio.io/istio/pkg/test/util/retry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKernelNativeModeRestart(t *testing.T) {
	framework.NewTest(t).Run(func(t framework.TestContext) {
		if !isKernelNativeMode() {
			t.Skipf("skipping kernel-native only test: mode=%s", currentKmeshMode())
		}

		assertKmeshDaemonMode(t, kernelNativeMode)

		restartKmesh(t)

		src := apps.EnrolledToKmesh[0]
		src.CallOrFail(t, echo.CallOptions{
			To:    apps.ServiceWithWaypointAtServiceGranularity,
			Count: 20,
			Port:  echo.Port{Name: "http"},
			Check: httpValidator,
		})
	})
}

func TestKernelNativeModeLargeScale(t *testing.T) {
	framework.NewTest(t).Run(func(t framework.TestContext) {
		if !isKernelNativeMode() {
			t.Skipf("skipping kernel-native only test: mode=%s", currentKmeshMode())
		}

		replicas := kernelNativeLargeScaleReplicaSize()
		t.Logf("deploying kernel-native large scale test workloads with %d replicas", replicas)

		builder := deployment.New(t).
			WithClusters(t.Clusters()...).
			WithConfig(echo.Config{
				Service:   "kernel-native-large-scale-server",
				Namespace: apps.Namespace,
				Ports:     ports.All(),
				Subsets: []echo.SubsetConfig{{
					Version:  "v1",
					Replicas: replicas,
				}},
			}).
			WithConfig(echo.Config{
				Service:   "kernel-native-large-scale-client",
				Namespace: apps.Namespace,
				Ports:     ports.All(),
				Subsets: []echo.SubsetConfig{{
					Version:  "v1",
					Replicas: 1,
				}},
			})

		echos, err := builder.Build()
		if err != nil {
			t.Fatal(err)
		}

		client := match.ServiceName(echo.NamespacedName{Name: "kernel-native-large-scale-client", Namespace: apps.Namespace}).GetMatches(echos)
		server := match.ServiceName(echo.NamespacedName{Name: "kernel-native-large-scale-server", Namespace: apps.Namespace}).GetMatches(echos)
		if len(client) == 0 || len(server) == 0 {
			t.Fatalf("failed to build large-scale workloads client=%d server=%d", len(client), len(server))
		}

		if len(server[0].WorkloadsOrFail(t)) != replicas {
			t.Fatalf("server workload size mismatch, got=%d want=%d", len(server[0].WorkloadsOrFail(t)), replicas)
		}

		callCount := replicas * 3
		client[0].CallOrFail(t, echo.CallOptions{
			To:    server,
			Count: callCount,
			Port:  echo.Port{Name: "tcp"},
			Check: check.OK(),
		})
	})
}

func assertKmeshDaemonMode(t framework.TestContext, expectedMode string) {
	daemonSetAPI := t.Clusters().Default().Kube().AppsV1().DaemonSets(KmeshNamespace)
	if err := retry.UntilSuccess(func() error {
		daemonSet, err := daemonSetAPI.Get(context.Background(), KmeshDaemonsetName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for _, container := range daemonSet.Spec.Template.Spec.Containers {
			if container.Name != "kmesh" {
				continue
			}
			args := strings.Join(container.Args, " ")
			if strings.Contains(args, fmt.Sprintf("--mode=%s", expectedMode)) {
				return nil
			}
			return fmt.Errorf("kmesh daemon args %q do not contain --mode=%s", args, expectedMode)
		}
		return fmt.Errorf("container kmesh not found in daemonset %s/%s", KmeshNamespace, KmeshDaemonsetName)
	}, retry.Timeout(2*time.Minute), retry.Delay(2*time.Second)); err != nil {
		t.Fatal(err)
	}
}
