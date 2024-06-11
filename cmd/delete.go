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
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/spf13/cobra"

	"github.com/shiftstack/cluster-dns/internal/client"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <cluster name>",
	Short: "Delete records for a cluster",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		client, err := client.GetRoute53Client(ctx, awsProfile)
		if err != nil {
			log.Fatal(err)
		}

		err = deleteClusterRecords(ctx, client, hostedZoneID, args[0])
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func deleteClusterRecords(ctx context.Context, client *route53.Client, hostedZoneID, clusterName string) error {
	recordsListOpts := &route53.ListResourceRecordSetsInput{
		HostedZoneId: &hostedZoneID,
	}

	var deleteRRS []*types.ResourceRecordSet
	var deleteNames []string

	for {
		records, err := client.ListResourceRecordSets(ctx, recordsListOpts)
		if err != nil {
			return err
		}

		for i := range records.ResourceRecordSets {
			rrs := &records.ResourceRecordSets[i]
			name := aws.ToString(rrs.Name)

			// Wildcard becomes "\052"
			if strings.HasPrefix(name, "api."+clusterName+".") || strings.HasPrefix(name, "\\052.apps."+clusterName+".") {
				deleteRRS = append(deleteRRS, rrs)
				deleteNames = append(deleteNames, name)
			}
		}

		if !records.IsTruncated {
			break
		}
		recordsListOpts.StartRecordIdentifier = records.NextRecordIdentifier
		recordsListOpts.StartRecordName = records.NextRecordName
		recordsListOpts.StartRecordType = records.NextRecordType
	}

	if len(deleteRRS) == 0 {
		log.Printf("No records found")
		return nil
	}

	change := route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{},
			Comment: aws.String("Delete records for cluster " + clusterName),
		},
		HostedZoneId: &hostedZoneID,
	}

	for _, rrs := range deleteRRS {
		change.ChangeBatch.Changes = append(change.ChangeBatch.Changes, types.Change{
			Action:            types.ChangeActionDelete,
			ResourceRecordSet: rrs,
		})
	}

	_, err := client.ChangeResourceRecordSets(ctx, &change)
	if err != nil {
		return err
	}

	log.Printf("Deleted: %s", strings.Join(deleteNames, " "))

	return nil
}
