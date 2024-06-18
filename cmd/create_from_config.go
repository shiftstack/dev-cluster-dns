/*
Copyright © 2024 Red Hat, Inc.

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
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/shiftstack/dev-cluster-dns/internal/client"
)

type Metadata struct {
	Name string `yaml:"name"`
}

type OpenStack struct {
	APIFloatingIP string `yaml:"apiFloatingIP"`
	IngressFloatingIP string `yaml:"ingressFloatingIP"`
}

type Platform struct {
	OpenStack OpenStack `yaml:"openstack"`
}

type InstallConfig struct {
	BaseDomain string `yaml:"baseDomain"`
	Metadata Metadata `yaml:"metadata"`
	Platform Platform `yaml:"platform"`
}

var createFromConfigCmd = &cobra.Command{
	Use:   "create-from-config <path>",
	Short: "Create or update records for a cluster from an install-config.yaml",
	Long: `Create or update DNS records for an OpenShift cluster in an existing HostedZone.

$ cluster-dns create-from-config install-config.yaml`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		client, err := client.GetRoute53Client(ctx, awsProfile)
		if err != nil {
			log.Fatal(err)
		}

		err = createClusterRecordsFromInstallConfig(ctx, client, args[0])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createFromConfigCmd)

	createFromConfigCmd.Flags().Int64Var(&createClusterRecordsOpts.TTL, "ttl", 60, "TTL of created records, in seconds")
	createFromConfigCmd.Flags().Int64Var(&createClusterRecordsOpts.WaitSeconds, "wait", 0, "Seconds to wait for records to be active. Set to zero to disable waiting")
}

func createClusterRecordsFromInstallConfig(ctx context.Context, client *route53.Client, path string) error {
		if _, err := os.Stat(path); err != nil {
			log.Fatal(err)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}

		installConfig := InstallConfig{}
		err = yaml.Unmarshal(data, &installConfig)
		if err != nil {
			log.Fatal(err)
		}

		// TODO: Support other zones. This will require a lookup.
		if hostedZoneID != shiftStackDevHostedZone {
			log.Fatal("Only the default hosted zone is currently supported")
		}

		if installConfig.BaseDomain != "shiftstack-dev.devcluster.openshift.com" {
			log.Fatal("baseDomain must be set to 'shiftstack-dev.devcluster.openshift.com'")
		}

		if installConfig.Metadata.Name == "" {
			log.Fatal("installConfig.Metadata.Name must be set")
		}

		if installConfig.Platform.OpenStack.APIFloatingIP == "" {
			log.Fatal("platform.openstack.apiFloatingIP must be set")
		}
		createClusterRecordsOpts.APIIP = installConfig.Platform.OpenStack.APIFloatingIP

		if installConfig.Platform.OpenStack.IngressFloatingIP == "" {
			log.Fatal("platform.openstack.ingressFloatingIP must be set")
		}
		createClusterRecordsOpts.IngressIP = installConfig.Platform.OpenStack.IngressFloatingIP

		err = createClusterRecords(ctx, client, hostedZoneID, installConfig.Metadata.Name, &createClusterRecordsOpts)
		if err != nil {
			log.Fatal(err)
		}

		return nil
}
