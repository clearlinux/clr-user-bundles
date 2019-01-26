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
	"github.com/spf13/cobra"
	"clr-user-bundles/swupd-wrapper/operations"
)

var updateCmd = &cobra.Command{
	Use: "update",
	Short: "Update 3rd party bundle content",
	Args: func(cmd *cobra.Command, args []string) error {
		if cmd.PersistentFlags().Changed("run-post") {
			runPost = false
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		operations.Update(StateDirectory, ContentDirectory, runPost)
	},
}

func init() {
	updateCmd.PersistentFlags().BoolVarP(&runPost, "run-post", "p", true, "Run post-3rd-party hook")
	rootCmd.AddCommand(updateCmd)
}
