// Copyright © 2019 Intel Corporation
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
	"fmt"
	"log"
	"github.com/spf13/cobra"
	"clr-user-bundles/post-job/operations"
)

var rootCmd = &cobra.Command{
	Long: `Process installed swupd 3rd party content into a usable form.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if StateDirectory[0] != '/' {
			return fmt.Errorf("statedir path must be absolute")
		}
		if ContentDirectory[0] != '/' {
			return fmt.Errorf("contentdir path must be absolute")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		operations.ProcessContent(StateDirectory, ContentDirectory)
	},
}

var StateDirectory string
var ContentDirectory string

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&StateDirectory, "statedir", "s", "/var/lib/swupd", "swupd state directory")
	rootCmd.PersistentFlags().StringVarP(&ContentDirectory, "contentdir", "c", "/var/lib/3rd-party", "3rd-party content directory")
}
