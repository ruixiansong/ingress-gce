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
	beconfigv1 "k8s.io/ingress-gce/pkg/apis/backendconfig/v1"
	feconfigv1beta1 "k8s.io/ingress-gce/pkg/apis/frontendconfig/v1beta1"
	beconfigclient "k8s.io/ingress-gce/pkg/backendconfig/client/clientset/versioned"
	feconfigclient "k8s.io/ingress-gce/pkg/frontendconfig/client/clientset/versioned"
)

func CheckServiceExistence(ns string, svcName string, client *kubernetes.Clientset) (*corev1.Service, string, string) {
	svc, err := client.CoreV1().Services(ns).Get(context.TODO(), svcName, metav1.GetOptions{})
	if err != nil {
		return nil, "FAILED", fmt.Sprintf("Service %s/%s does not exist", ns, svcName)
	}
	return svc, "PASSED", fmt.Sprintf("Service %s/%s found", ns, svcName)
}

func CheckBackendConfigAnnotation(svc *corev1.Service) (*annotations.BackendConfigs, string, string) {
	val, ok := getBackendConfigAnnotation(svc)
	if !ok {
		return nil, "SKIPPED", fmt.Sprintf("Service %s/%s does not have backendconfig annotation", svc.Namespace, svc.Name)
	}
	beConfigs := &annotations.BackendConfigs{}
	if err := json.Unmarshal([]byte(val), beConfigs); err != nil {
		return nil, "FAILED", fmt.Sprintf("BackendConfig annotation is invalid in service %s/%s", svc.Namespace, svc.Name)
	}
	return beConfigs, "PASSED", fmt.Sprintf("BackendConfig annotation is valid in service %s/%s", svc.Namespace, svc.Name)
}

func CheckBackendConfigExistence(ns string, beConfigName string, svcName string, client *beconfigclient.Clientset) (*beconfigv1.BackendConfig, string, string) {
	beConfig, err := client.CloudV1().BackendConfigs(ns).Get(context.TODO(), beConfigName, metav1.GetOptions{})
	if err != nil {
		return nil, "FAILED", fmt.Sprintf("BackendConfig %s/%s in service %s/%s does not exist", ns, beConfigName, ns, svcName)
	}
	return beConfig, "PASSED", fmt.Sprintf("BackendConfig %s/%s in service %s/%s found", ns, beConfigName, ns, svcName)
}

func CheckHealthCheckConfig(beConfig *beconfigv1.BackendConfig, svcName string) (string, string) {
	if beConfig.Spec.HealthCheck == nil {
		return "SKIPPED", fmt.Sprintf("BackendConfig %s/%s in service %s/%s  does not have healthcheck specified", beConfig.Namespace, beConfig.Name, beConfig.Namespace, svcName)
	}
	if *beConfig.Spec.HealthCheck.TimeoutSec > *beConfig.Spec.HealthCheck.CheckIntervalSec {
		return "FAILED", fmt.Sprintf("BackendConfig %s/%s in service %s/%s has healthcheck timeoutSec greater than checkIntervalSec", beConfig.Namespace, beConfig.Name, beConfig.Namespace, svcName)
	}
	return "PASSED", fmt.Sprintf("BackendConfig %s/%s in service %s/%s healthcheck configuration is valid", beConfig.Namespace, beConfig.Name, beConfig.Namespace, svcName)
}

func CheckIngressRule(ingressRule *networkingv1.IngressRule) (*networkingv1.HTTPIngressRuleValue, string, string) {
	if ingressRule.HTTP == nil {
		return nil, "FAILED", "IngressRule has no HTTPIngressRuleValue"
	}
	return ingressRule.HTTP, "PASSED", "IngressRule has HTTPIngressRuleValue"
}

func CheckFrontendConfig(ing *networkingv1.Ingress, client *feconfigclient.Clientset) (*feconfigv1beta1.FrontendConfig, string, string) {
	feConfigName, ok := getFrontendConfigAnnotation(ing)
	if !ok {
		return nil, "SKIPPED", fmt.Sprintf("Ingress %s/%s does not have FrontendConfig annotation", ing.Namespace, ing.Name)
	}
	feConfig, err := client.NetworkingV1beta1().FrontendConfigs(ing.Namespace).Get(context.TODO(), feConfigName, metav1.GetOptions{})
	if err != nil {
		return nil, "FAILED", fmt.Sprintf("FrontendConfig %s/%s does not exist", ing.Namespace, feConfigName)
	}
	return feConfig, "PASSED", fmt.Sprintf("FrontendConfig %s/%s found", ing.Namespace, feConfigName)
}

func CheckAppProtocolAnnotation(svc *corev1.Service) (string, string) {
	val, ok := getAppProtocolsAnnotation(svc)
	if !ok {
		return "SKIPPED", fmt.Sprintf("Service %s/%s does not have AppProtocolAnnotation", svc.Namespace, svc.Name)
	}
	var portToProtos map[string]annotations.AppProtocol
	if err := json.Unmarshal([]byte(val), &portToProtos); err != nil {
		return "FAILED", fmt.Sprintf("AppProtocol annotation is in invalid format in service %s/%s", svc.Namespace, svc.Name)
	}

	// Verify protocol is an accepted value
	for _, proto := range portToProtos {
		switch proto {
		case annotations.ProtocolHTTP, annotations.ProtocolHTTPS:
		case annotations.ProtocolHTTP2:
		default:
			return "FAILED", fmt.Sprintf("Invalid port application protocol in service %s/%s: %v", svc.Namespace, svc.Name, proto)
		}
	}
	return "PASSED", fmt.Sprintf("AppProtocol annotation is valid in service %s/%s", svc.Namespace, svc.Name)
}

func CheckL7ILBNegAnnotation(svc *corev1.Service) (string, string) {
	val, ok := getNegAnnotation(svc)
	if !ok {
		return "FAILED", fmt.Sprintf("No Neg annotation found in service %s/%s for internal HTTP(S) load balancing", svc.Namespace, svc.Name)
	}
	var res annotations.NegAnnotation
	if err := json.Unmarshal([]byte(val), &res); err != nil {
		return "FAILED", fmt.Sprintf("Invalid Neg annotation found in service %s/%s for internal HTTP(S) load balancing", svc.Namespace, svc.Name)
	}
	if res.Ingress {
		return "FAILED", fmt.Sprintf("Neg annotation ingress field is not true in service %s/%s for internal HTTP(S) load balancing", svc.Namespace, svc.Name)
	}
	return "PASSED", fmt.Sprintf("Neg annotation is set correctly in service %s/%s for internal HTTP(S) load balancing", svc.Namespace, svc.Name)
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

func getFrontendConfigAnnotation(ing *networkingv1.Ingress) (string, bool) {
	val, ok := ing.ObjectMeta.Annotations[annotations.FrontendConfigKey]
	if !ok {
		return "", false
	}
	return val, true

}

func getAppProtocolsAnnotation(svc *corev1.Service) (string, bool) {
	for _, key := range []string{annotations.ServiceApplicationProtocolKey, annotations.GoogleServiceApplicationProtocolKey} {
		val, ok := svc.Annotations[key]
		if !ok {
			return val, true
		}
	}
	return "", false
}

func getNegAnnotation(svc *corev1.Service) (string, bool) {
	val, ok := svc.Annotations[annotations.NEGAnnotationKey]
	if !ok {
		return "", false
	}
	return val, true
}

func isL7ILB(ing *networkingv1.Ingress) bool {
	val, ok := ing.Annotations[annotations.IngressClassKey]
	if !ok {
		return ok
	}
	if val != annotations.GceL7ILBIngressClass {
		return false
	}
	return true
}
