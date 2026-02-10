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
	"os"
	"strconv"
	"strings"
)

const (
	kmeshModeEnvVar                = "KMESH_E2E_KMESH_MODE"
	defaultKmeshMode               = "dual-engine"
	kernelNativeMode               = "kernel-native"
	kernelNativeLargeScaleEnvVar   = "KMESH_E2E_KERNEL_NATIVE_LARGE_SCALE_REPLICAS"
	defaultKernelNativeReplicaSize = 20
)

func currentKmeshMode() string {
	mode := strings.TrimSpace(strings.ToLower(os.Getenv(kmeshModeEnvVar)))
	if mode == "" {
		return defaultKmeshMode
	}
	return mode
}

func isKernelNativeMode() bool {
	return currentKmeshMode() == kernelNativeMode
}

func kernelNativeLargeScaleReplicaSize() int {
	replicas, err := strconv.Atoi(strings.TrimSpace(os.Getenv(kernelNativeLargeScaleEnvVar)))
	if err != nil || replicas <= 0 {
		return defaultKernelNativeReplicaSize
	}
	return replicas
}
