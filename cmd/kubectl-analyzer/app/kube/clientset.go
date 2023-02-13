// Copyright 2023 the Kubernetes Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	beconfigclient "k8s.io/ingress-gce/pkg/backendconfig/client/clientset/versioned"
	feconfigclient "k8s.io/ingress-gce/pkg/frontendconfig/client/clientset/versioned"
)

// NewClientSet returns a new Kubernetes clientset
func NewClientSet(kubeContext, kubeConfig string) (*kubernetes.Clientset, error) {
	config, err := getKubeConfig(kubeContext, kubeConfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func getKubeConfig(kubeContext, kubeConfig string) (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeConfig != "" {
		loadingRules.ExplicitPath = kubeConfig
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{CurrentContext: kubeContext},
	).ClientConfig()
}

func NewBackendConfigClientSet(kubeContext, kubeConfig string) (*beconfigclient.Clientset, error) {
	config, err := getKubeConfig(kubeContext, kubeConfig)
	if err != nil {
		return nil, err
	}

	return beconfigclient.NewForConfig(config)
}

func NewFrontendConfigClientSet(kubeContext, kubeConfig string) (*feconfigclient.Clientset, error) {
	config, err := getKubeConfig(kubeContext, kubeConfig)
	if err != nil {
		return nil, err
	}

	return feconfigclient.NewForConfig(config)
}
