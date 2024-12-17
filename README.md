# Testing and developing the `f5-virtualserver` source in `external-dns`

This repository can be used as inspiration for adding a new `external-dns` source, basically a source to use when creating DNS records in a specific provider e.g. AWS (Route 53) or Designate (OpenStack).

In this proof-of-concept i've created a new source to create DNS records based on F5 Networks `VirtualServer` CRDs. There's two fields in the `VirtualServer` CRD that is of interest, the `host` and the `virtualServerAddress` fields.

To test this out i build `external-dns` locally and create a `kind` cluster, when i start `external-dns` locally it'll connect to the `kind` cluster and start to watch for `VirtualServer` CRDs.

One awesome feature in `external-dns` is that you can use the `inmemory` provider to store DNS records in memory, you don't need something live to connect to. Useful to do some manual testing of `external-dns`.

_Please note that the `f5-virtualserver` source will enumerate all `VirtualServers` in the cluster, some virtual servers will have a static IP address assigned others through an IPAM controller that writes the IP address in the status field of the `VirtualServer`. The `ipam-controller` directory in this repository includes code to do exactly that, write to the status field. This code can be used to test both scenarios._

## Step-by-step

1. Clone your `external-dns` fork.

2. Checkout your feature branch.

3. Create the `kind` cluster:

```bash
kind create cluster --config=dev-cluster.yaml
```

4. Install the CRDs shipped from F5 Networks (defined as part of the `k8s-bigip-ctlr` repository):

```bash
kubectl create -f https://raw.githubusercontent.com/F5Networks/k8s-bigip-ctlr/refs/heads/master/docs/cis-20.x/config_examples/customResourceDefinitions/stable/customresourcedefinitions.yml
```

5. Build `external-dns` locally:

```bash
make build
```

## Testing with a `VirtualServer` with a static IP address configured

1. Start `external-dns`:

```bash
./build/external-dns \
  --source=f5-virtualserver \
  --provider=inmemory \
  --log-level=debug \
  --policy=upsert-only \
  --registry=txt \
  --interval=1m \
  --txt-owner-id=external-dns-cluster \
  --domain-filter=example.com
```

2. Create the `VirtualServer` object:

```bash
kubectl create -f virtualserver-static.yaml
```

3. See the logs of `external-dns`.

## Testing with a `VirtualServer` with a dynamically configured IP address (via the included fake `ipam-controller`)

1. Start `external-dns`:

```bash
./build/external-dns \
  --source=f5-virtualserver \
  --provider=inmemory \
  --log-level=debug \
  --policy=upsert-only \
  --registry=txt \
  --interval=1m \
  --txt-owner-id=external-dns-cluster \
  --domain-filter=example.com
```

2. Create the `VirtualServer` object:

```bash
kubectl create -f virtualserver-ipam.yaml
```

3. Run the fake `ipam-controller`:

```bash
cd ipam-controller
go run main.go
```

This will update the `status` field of the `VirtualServer` called `example-vs-ipam` in the `default` namespace.

### 2024-12-17

We're not handling cases when the `status.status` field of the `VirtualServer` is not `Ok` e.g. `""` or `ERROR` which means that we don't exit early in the `f5-virtualserver` source. If we end up with a status of e.g. `ERROR` `external-dns` will still try to create a record AND upsert (basically removing the old one), the record it tries to create is of type CNAME, which is valid and it makes sense.

## Testing with the `DNSEndpoint` CRD

You'll see `external-dns` create records in the `inmemory` provider by watching the standard output of the `external-dns` binary. If you want to test the `crd` source, which watches for `DNSEndpoint` CRDs (provided by `external-dns`):

1. Install the CRD:

```bash
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/external-dns/master/docs/contributing/crd-source/crd-manifest.yaml
```

2. Create a `DNSEndpoint` object:

```bash
kubectl create -f dnsendpoint.yaml
```
