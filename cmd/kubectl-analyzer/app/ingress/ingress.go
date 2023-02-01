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
	beconfigClient, err := kube.NewBackendConfigClientSet(kubecontext, kubeconfig)
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

		/*
			#TODO: add checks for ingress and frontendConfig
		*/

		// Get all the services specified in the ingress and add service names to a list.
		svcNameList := []string{}
		if ingress.Spec.DefaultBackend != nil {
			svcNameList = append(svcNameList, ingress.Spec.DefaultBackend.Service.Name)
		}
		if ingress.Spec.Rules != nil {
			for _, rule := range ingress.Spec.Rules {
				httpIngressRule, res, msg := CheckIngressRule(&rule)
				ingressRes.Checks = append(ingressRes.Checks, &util.Check{
					Id:  "IngressRuleCheck",
					Msg: msg,
					Res: res,
				})
				if httpIngressRule != nil {
					for _, path := range rule.HTTP.Paths {
						if path.Backend.Service != nil {
							svcNameList = append(svcNameList, path.Backend.Service.Name)
						}
					}
				}
			}
		}
		for _, svcName := range svcNameList {
			svc, res, msg := CheckServiceExistence(ingress.Namespace, svcName, client)
			ingressRes.Checks = append(ingressRes.Checks, &util.Check{
				Id:  "ServiceExistenceCheck",
				Msg: msg,
				Res: res,
			})

			// If service exists, go ahead and check other service rules.
			if svc != nil {
				/*
					#TODO: add checks for other service annotations
				*/
				beConfigs, res, msg := CheckBackendConfigAnnotation(svc)
				ingressRes.Checks = append(ingressRes.Checks, &util.Check{
					Id:  "BackendConfigAnnotationCheck",
					Msg: msg,
					Res: res,
				})
				// If backendConfig annotation is valid, go ahead and check other backendConfig rules.
				if beConfigs != nil {
					// Get all the backendconfigs in the annotation and add backendconfig names to a list.
					beConfigNameList := []string{}
					if beConfigs.Default != "" {
						beConfigNameList = append(beConfigNameList, beConfigs.Default)
					}
					for _, beConfigName := range beConfigs.Ports {
						beConfigNameList = append(beConfigNameList, beConfigName)
					}
					for _, beConfigName := range beConfigNameList {
						beConfig, res, msg := CheckBackendConfigExistence(ingress.Namespace, beConfigName, beconfigClient)
						ingressRes.Checks = append(ingressRes.Checks, &util.Check{
							Id:  "BackendConfigExistenceCheck",
							Msg: msg,
							Res: res,
						})
						// If backendConfig exists, go ahead and check other backendConfig rules.
						if beConfig != nil {
							res, msg := CheckHealthCheckConfig(beConfig)
							ingressRes.Checks = append(ingressRes.Checks, &util.Check{
								Id:  "HealthCheckConfigCheck",
								Msg: msg,
								Res: res,
							})
						}
					}
				}
			}
		}
		res.Ingress = append(res.Ingress, ingressRes)
	}
	util.JsonReport(res)
}
