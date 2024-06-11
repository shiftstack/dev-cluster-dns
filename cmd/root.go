/*
Copyright © 2024 Matthew Booth

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	shiftStackDevHostedZone = "Z0400818H9HMCRQLQP0V"
	defaultAWSProfile       = "saml"
)

var (
	cfgFile      string
	hostedZoneID string
	awsProfile   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dev-cluster-dns",
	Short: "Manage DNS records for an OpenShift cluster in an existing Route 53 Hosted Zone",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.dev-cluster-dns.yaml)")

	rootCmd.PersistentFlags().StringVar(&hostedZoneID, "hosted-zone", shiftStackDevHostedZone, "ID of the HostedZone where records will be created")
	viper.BindPFlag("hosted-zone", rootCmd.Flags().Lookup("hosted-zone"))

	rootCmd.PersistentFlags().StringVar(&awsProfile, "aws-profile", defaultAWSProfile, "AWS credentials profile to read from")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// XXX(mdbooth): viper integration does not appear to be working
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".dev-cluster-dns" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("toml")
		viper.SetConfigName(".dev-cluster-dns")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
