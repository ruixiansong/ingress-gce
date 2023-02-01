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
)

const (
	//StandardOutput is the constant value for output type standard
	StandardOutput string = "standard"
	//JSONOutput is the constant value for output type JSON
	JSONOutput string = "json"
)

type Report struct {
	Ingress []*Ingress `json:"ingress,omitempty" `
}

type Ingress struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Checks    []*Check `json:"checks,omitempty"`
}

type Check struct {
	Id  string `json:"id"`
	Msg string `json:"msg"`
	Res string `json:"res"`
}

func JsonReport(report Report) {
	jsonRaw, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Println("Error Marshalling JSON")
		fmt.Println(err)
	}
	fmt.Printf("%s", jsonRaw)
}

// SupportedOutputs returns a string list of output formats supposed by this package
func SupportedOutputs() []string {
	return []string{
		StandardOutput,
		JSONOutput,
	}
}
