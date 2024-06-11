# dev-cluster-dns

Creates and deletes DNS records for an OpenShift cluster in AWS route 53. Specifically it creates:

- api.&lt;_cluster name_&gt;.&lt;_hosted zone_&gt;
- *.apps.&lt;_cluster name_&gt;.&lt;_hosted zone_&gt;

By default it uses zone `Z0400818H9HMCRQLQP0V`, which is `shiftstack-dev.devcluster.openshift.com`. This can be overridden with the `--hosted-zone` option.

## Create records

```
$ dev-cluster-dns create my-cluster --api 192.0.2.1 --ingress 192.0.2.2
2024/06/11 14:24:20 Base domain: shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:24:20 Create or update records: api.my-cluster.shiftstack-dev.devcluster.openshift.com. *.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:24:20 Status: PENDING
```

Note that the base domain of the hosted cluster is added automatically.

Note that the status is `PENDING`, meaning that the records are not published yet. Use the `--wait` flag if you need to wait until the records are published. A recommended value is 120 seconds.

The default TTL for the records is 60 seconds. It can be changed with the `--ttl` flag.

`create` will also update existing records.

## Delete records

```
$ dev-cluster-dns delete my-cluster
2024/06/11 14:28:00 Deleted: api.my-cluster.shiftstack-dev.devcluster.openshift.com. \052.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
```