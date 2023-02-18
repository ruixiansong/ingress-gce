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

package util

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	//JSONOutput is the constant value for output type JSON
	JSONOutput string = "json"
)

// Report represents the final output of the analyzer
type Report struct {
	Resource []*Resource `json:"ingress,omitempty"`
	Error    []string    `json:"error,omitempty"`
}

// Resource represents the a resource of the cluster and all the checks done on it
type Resource struct {
	Type      string   `json:"type"`
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Checks    []*Check `json:"checks,omitempty"`
}

// Check represents the result of a check
type Check struct {
	Name string `json:"name"`
	Msg  string `json:"msg"`
	Res  string `json:"res"`
}

func JsonReport(report Report) string {
	jsonRaw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("Error Marshalling JSON: %s\n", jsonRaw)
		os.Exit(1)
	}
	return string(jsonRaw)
}

// SupportedOutputs returns a string list of output formats supposed by this package
func SupportedOutputs() []string {
	return []string{
		JSONOutput,
	}
}
