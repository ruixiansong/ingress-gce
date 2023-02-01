/*
Copyright 2023 The Kubernetes Authors.

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

package ingress

import (
	"context"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/ingress-gce/cmd/kubectl-analyzer/app/kube"
	"k8s.io/ingress-gce/cmd/kubectl-analyzer/app/util"
)

func CheckAllIngresses(kubeconfig, kubecontext, namespace string) {
	client, err := kube.NewClientSet(kubecontext, kubeconfig)
	if err != nil {
		fmt.Printf("Error connecting to Kubernetes: %v\n", err)
	}

	ingressList, err := client.NetworkingV1().Ingresses(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing ingresses: %v\n", err)
		os.Exit(1)
	}
	res := util.Report{
		Ingress: []*util.Ingress{},
	}
	for _, ingress := range ingressList.Items {
		ingressRes := &util.Ingress{
			Namespace: ingress.Namespace,
			Name:      ingress.Name,
			Checks:    []*util.Check{},
		}
		res.Ingress = append(res.Ingress, ingressRes)
	}
	util.JsonReport(res)
}
