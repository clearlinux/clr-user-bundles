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
	"github.com/spf13/cobra"
	"clr-user-bundles/swupd-wrapper/operations"
)

var removeCmd = &cobra.Command{
	Use: "remove [URI to 3rd party content] [BUNDLE-NAME]",
	Short: "Remove 3rd party bundle content",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("Invalid arguments")
		}
		if cmd.PersistentFlags().Changed("skip-post") {
			skipPost = true
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		operations.Remove(StateDirectory, ContentDirectory, args[0], args[1], skipPost)
	},
}

func init() {
	removeCmd.PersistentFlags().BoolVarP(&skipPost, "skip-post", "p", false, "Skip running post-3rd-party hooks")
	rootCmd.AddCommand(removeCmd)
}
