// Copyright 2018 Istio Authors
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

package cmd

import (
	"github.com/spf13/cobra"
	"istio.io/istio/galley/cmd/shared"
	"istio.io/istio/galley/pkg/config/sync"
	"istio.io/istio/pkg/log"
)

func initCmd(rootArgs *rootArgs, fatalf shared.FormatFn) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Do a one-time copy of CRDs.",
		Long:  "Co a one-time copy of CRDs.",

		Run: func(cmd *cobra.Command, args []string) {
			err := initCrds(rootArgs)
			if err != nil {
				fatalf("%v", err)
			}
		}}
}

func initCrds(rootArgs *rootArgs) (err error) {
	if err = log.Configure(rootArgs.loggingOptions); err != nil {
		return
	}

	var controller *sync.Controller
	if controller, err = sync.NewController(rootArgs.kubeconfigPath); err != nil {
		return
	}

	return controller.CopyCRDs()
}
