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

package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/ingress-gce/cmd/gke-diagnoser/app/util"
)

var kubeconfig string
var kubecontext string
var namespace string
var outputFormat string

var rootCmd = &cobra.Command{
	Use:   "kubectl check-gke-ingress",
	Short: "kubectl check-gke-ingress is a kubectl tool to check the correctness of ingress and ingress related resources.",
	Long:  "kubectl check-gke-ingress is a kubectl tool to check the correctness of ingress and ingress related resources.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.ParseFlags(args); err != nil {
			fmt.Printf("Error parsing flags: %v", err)
		}
		if err := validateOutputType(outputFormat); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("Starting check-gke-ingress")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "kubeconfig file to use for Kubernetes config")
	rootCmd.PersistentFlags().StringVarP(&kubecontext, "context", "c", "", "context to use for Kubernetes config")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "only check resources from this namespace")
	rootCmd.PersistentFlags().StringVarP(&outputFormat,
		"output", "o", util.JSONOutput,
		fmt.Sprintf("output format for check results (supports: %v)", util.SupportedOutputs()))
}

// Execute is the primary entrypoint for this CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func validateOutputType(outputType string) error {
	for _, format := range util.SupportedOutputs() {
		if format == outputType {
			return nil
		}
	}
	return fmt.Errorf("Unsupported Output Type. We only support: %v", util.SupportedOutputs())
}
