# dev-cluster-dns

## Usage

Creates and deletes DNS records for an OpenShift cluster in AWS route 53. Specifically it creates:

- api.&lt;_cluster name_&gt;.&lt;_hosted zone_&gt;
- api-int.&lt;_cluster name_&gt;.&lt;_hosted zone_&gt; (when using `--host`)
- *.apps.&lt;_cluster name_&gt;.&lt;_hosted zone_&gt;

By default it uses zone `Z0400818H9HMCRQLQP0V`, which is `shiftstack-dev.devcluster.openshift.com`. This can be overridden with the `--hosted-zone` option.

### Create records

To create or update A records for a cluster manually:

```
$ dev-cluster-dns create my-cluster --api 192.0.2.1 --ingress 192.0.2.2
2024/06/11 14:24:20 Base domain: shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:24:20 Create or update records: api.my-cluster.shiftstack-dev.devcluster.openshift.com. *.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:24:20 Status: PENDING
```

To create or update CNAME records pointing to a single host (e.g. for [SNO][sno]):

```
$ dev-cluster-dns create my-cluster --host my-host.example.com
2024/06/11 14:24:20 Base domain: shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:24:20 Create or update records: api.my-cluster.shiftstack-dev.devcluster.openshift.com. api-int.my-cluster.shiftstack-dev.devcluster.openshift.com. *.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:24:20 Status: PENDING
```

Alternatively, to create or update from an `install-config.yaml` file:

```
$ dev-cluster-dns create-from-config install-config.yaml
2024/06/11 14:36:01 Base domain: shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:36:01 Create or update records: api.my-cluster.shiftstack-dev.devcluster.openshift.com. *.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
2024/06/11 14:36:01 Status: PENDING
```

Note that the base domain of the hosted cluster is added automatically.

Note that the status is `PENDING`, meaning that the records are not published yet. Use the `--wait` flag if you need to wait until the records are published. A recommended value is 120 seconds.

The default TTL for the records is 60 seconds. It can be changed with the `--ttl` flag.

[sno]: https://docs.redhat.com/en/documentation/openshift_container_platform/4.22/html/installing_on_a_single_node/index

### List records

```
$ dev-cluster-dns list
name: api.my-cluster.shiftstack-dev.devcluster.openshift.com.
        value: 192.0.2.1
name: *.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
        value: 192.0.2.2
```

### Delete records

```
$ dev-cluster-dns delete my-cluster
2024/06/11 14:28:00 Deleted: api.my-cluster.shiftstack-dev.devcluster.openshift.com. \052.apps.my-cluster.shiftstack-dev.devcluster.openshift.com.
```

## Building

This repo is configured to automatically generate a release whenever a tag is pushed. To create a new release, simply push a new tag.
