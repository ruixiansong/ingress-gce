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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func CheckServiceExistence(ns string, svcName string, client *kubernetes.Clientset) (*corev1.Service, string, string) {
	svc, err := client.CoreV1().Services(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return nil, "FAILED", fmt.Sprintf("ServiceExistence check FAILED, Service %s/%s does not exist", ns, svcName)
	}
	return svc, "PASSED", fmt.Sprintf("ServiceExistence check PASSED, Service %s/%s found", ns, svcName)
}
