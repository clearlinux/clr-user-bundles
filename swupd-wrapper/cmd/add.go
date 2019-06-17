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
	"github.com/clearlinux/clr-user-bundles/swupd-wrapper/operations"
)

var addCmd = &cobra.Command{
	Use: "add [URI to 3rd party content]",
	Short: "Add 3rd party bundle content",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Invalid URI specification")
		}
		if cmd.PersistentFlags().Changed("skip-post") {
			skipPost = true
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		operations.Add(args[0], StateDirectory, ContentDirectory, skipPost)
	},
}

func init() {
	addCmd.PersistentFlags().BoolVarP(&skipPost, "skip-post", "p", false, "Skip running post-3rd-party hooks")
	rootCmd.AddCommand(addCmd)
}
