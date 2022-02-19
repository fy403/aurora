/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"aurora/internal/center"

	"github.com/spf13/cobra"
)

// centerCmd represents the version command
var centerCmd = &cobra.Command{
	Use:   "center",
	Short: "start a center instance",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return center.Run()
	},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return center.PreRun()
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		return center.PostRun()
	},
}

func init() {
	// add subCmd to rootCmd Usage: aurora version
	rootCmd.AddCommand(centerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// centerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// centerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
