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
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/shiftstack/dev-cluster-dns/internal/client"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List existing records",
	Long:  `List DNS records for all OpenShift clusters in an existing HostedZone.

$ cluster-dns list`,
	Args: nil,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		client, err := client.GetRoute53Client(ctx, awsProfile)
		if err != nil {
			log.Fatal(err)
		}

		err = listClusterRecords(ctx, client, hostedZoneID)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func listClusterRecords(ctx context.Context, client *route53.Client, hostedZoneID string) error {
	filter := route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneID,
	}

	resourceRecordSets, err := client.ListResourceRecordSets(ctx, &filter)
	if err != nil {
		return err
	}

	for _, resourceRecordSet := range resourceRecordSets.ResourceRecordSets {
		name := aws.ToString(resourceRecordSet.Name)
		// The wildcard is ASCII encoded
		// https://github.com/fog/fog/issues/1093
		name = strings.Replace(name, "\\052", "*", 1)
		if strings.HasPrefix(name, "api.") || strings.HasPrefix(name, "*.apps.") {
			fmt.Printf("name: %s\n", name)
			for _, resourceRecord := range resourceRecordSet.ResourceRecords {
				fmt.Printf("\tvalue: %s\n", *resourceRecord.Value)
			}
		}
	}

	return nil
}
