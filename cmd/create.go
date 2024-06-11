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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/spf13/cobra"

	"github.com/shiftstack/cluster-dns/internal/client"
)

var (
	createClusterRecordsOpts CreateClusterRecordsOpts
)

type CreateClusterRecordsOpts struct {
	APIIP       string
	IngressIP   string
	TTL         int64
	WaitSeconds int64
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <cluster name>",
	Short: "Create or update records for a cluster",
	Long: `Create or update DNS records for an OpenShift cluster in an existing HostedZone.

$ cluster-dns create my-cluster --api 1.2.3.4 --ingress 1.2.3.5`,
	Args: cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO()

		client, err := client.GetRoute53Client(ctx, awsProfile)
		if err != nil {
			log.Fatal(err)
		}

		err = createClusterRecords(ctx, client, hostedZoneID, args[0], &createClusterRecordsOpts)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&createClusterRecordsOpts.APIIP, "api", "", "IP address of the api server endpoint")
	createCmd.Flags().StringVar(&createClusterRecordsOpts.IngressIP, "ingress", "", "IP address of the ingress endpoint")
	createCmd.MarkFlagsOneRequired("api", "ingress")

	createCmd.Flags().Int64Var(&createClusterRecordsOpts.TTL, "ttl", 60, "TTL of created records, in seconds")
	createCmd.Flags().Int64Var(&createClusterRecordsOpts.WaitSeconds, "wait", 0, "Seconds to wait for records to be active. Set to zero to disable waiting")
}

func createClusterRecords(ctx context.Context, client *route53.Client, hostedZoneID, clusterName string, opts *CreateClusterRecordsOpts) error {
	hostedZoneOutput, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: &hostedZoneID,
	})
	if err != nil {
		return err
	}

	baseDomain := aws.ToString(hostedZoneOutput.HostedZone.Name)
	log.Printf("Base domain: %s", baseDomain)

	change := route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{},
			Comment: aws.String("Create or update records for cluster " + clusterName),
		},
		HostedZoneId: &hostedZoneID,
	}

	var records []string
	if opts.APIIP != "" {
		name := "api." + clusterName + "." + baseDomain
		records = append(records, name)
		change.ChangeBatch.Changes = append(change.ChangeBatch.Changes, genClusterRecord(name, opts.TTL, opts.APIIP))
	}

	if opts.IngressIP != "" {
		name := "*.apps." + clusterName + "." + baseDomain
		records = append(records, name)
		change.ChangeBatch.Changes = append(change.ChangeBatch.Changes, genClusterRecord(name, opts.TTL, opts.IngressIP))
	}

	log.Printf("Create or update records: %s", strings.Join(records, " "))

	out, err := client.ChangeResourceRecordSets(ctx, &change)
	if err != nil {
		return err
	}

	log.Printf("Status: %s", out.ChangeInfo.Status)

	if opts.WaitSeconds > 0 {
		log.Printf("Waiting up to %d seconds for records to be updated", opts.WaitSeconds)

		waiter := route53.NewResourceRecordSetsChangedWaiter(client)
		output, err := waiter.WaitForOutput(ctx, &route53.GetChangeInput{
			Id: out.ChangeInfo.Id,
		}, time.Second*time.Duration(opts.WaitSeconds))
		if err != nil {
			return err
		}

		log.Printf("Status: %s", output.ChangeInfo.Status)
	}

	return nil
}

func genClusterRecord(name string, ttl int64, ip string) types.Change {
	return types.Change{
		Action: types.ChangeActionUpsert,
		ResourceRecordSet: &types.ResourceRecordSet{
			Name: &name,
			Type: types.RRTypeA,
			TTL:  &ttl,
			ResourceRecords: []types.ResourceRecord{
				{
					Value: &ip,
				},
			},
		},
	}

}
