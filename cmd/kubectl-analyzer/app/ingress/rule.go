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
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/ingress-gce/pkg/annotations"
	v1 "k8s.io/ingress-gce/pkg/apis/backendconfig/v1"
	beconfigclient "k8s.io/ingress-gce/pkg/backendconfig/client/clientset/versioned"
)

func CheckServiceExistence(ns string, svcName string, client *kubernetes.Clientset) (*corev1.Service, string, string) {
	svc, err := client.CoreV1().Services(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return nil, "FAILED", fmt.Sprintf("ServiceExistence check FAILED, Service %s/%s does not exist", ns, svcName)
	}
	return svc, "PASSED", fmt.Sprintf("ServiceExistence check PASSED, Service %s/%s found", ns, svcName)
}

func CheckBackendConfigAnnotation(svc *corev1.Service) (*annotations.BackendConfigs, string, string) {
	val, ok := getBackendConfigAnnotation(svc)
	if !ok {
		return nil, "SKIPPED", fmt.Sprintf("BackendConfigAnnotation check SKIPPED, Service %s/%s does not have backendconfig annotation", svc.Namespace, svc.Name)
	}
	beConfigs := &annotations.BackendConfigs{}
	if err := json.Unmarshal([]byte(val), beConfigs); err != nil {
		return nil, "FAILED", fmt.Sprintf("BackendConfigAnnotation check FAILED, BackendConfig annotation is invalid json in service %s/%s", svc.Namespace, svc.Name)
	}
	return beConfigs, "PASSED", fmt.Sprintf("BackendConfigAnnotation check PASSED, BackendConfig annotation is valid in service %s/%s", svc.Namespace, svc.Name)
}

func CheckBackendConfigExistence(ns string, name string, client *beconfigclient.Clientset) (*v1.BackendConfig, string, string) {
	beConfig, err := client.CloudV1().BackendConfigs(ns).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, "FAILED", fmt.Sprintf("BackendConfigExistence check FAILED, BackendConfig %s/%s does not exist", ns, name)
	}
	return beConfig, "PASSED", fmt.Sprintf("BackendConfigExistence check PASSED, BackendConfig %s/%s found", ns, name)
}

func CheckHealthCheckConfig(beConfig *v1.BackendConfig) (string, string) {
	if beConfig.Spec.HealthCheck == nil {
		return "SKIPPED", fmt.Sprintf("HealthCheckConfig check SKIPPED, BackendConfig %s/%s does not have healthcheck specified", beConfig.Namespace, beConfig.Name)
	}
	if *beConfig.Spec.HealthCheck.TimeoutSec > *beConfig.Spec.HealthCheck.CheckIntervalSec {
		return "FAILED", fmt.Sprintf("HealthCheckConfig check FAILED, BackendConfig %s/%s has healthcheck timeoutSec greater than checkIntervalSec", beConfig.Namespace, beConfig.Name)
	}
	return "PASSED", fmt.Sprintf("HealthCheckConfig check PASSED, BackendConfig %s/%s healthcheck configuration is valid", beConfig.Namespace, beConfig.Name)
}

func CheckIngressRule(ingressRule *networkingv1.IngressRule) (*networkingv1.HTTPIngressRuleValue, string, string) {
	if ingressRule.HTTP == nil {
		return nil, "FAILED", "IngressRule Check FAILED, IngressRule has no HTTPIngressRuleValue"
	}
	return ingressRule.HTTP, "PASSED", "IngressRule Check PASSED, IngressRule has HTTPIngressRuleValue"
}

func getBackendConfigAnnotation(svc *corev1.Service) (string, bool) {
	for _, bcKey := range []string{annotations.BackendConfigKey, annotations.BetaBackendConfigKey} {
		val, ok := svc.Annotations[bcKey]
		if ok {
			return val, ok
		}
	}
	return "", false
}
