# Proof-of-concept: Add a new external-dns source

This repository can be used as inspiration for adding a new `external-dns` source, basically a source to use when creating DNS records in a specific provider e.g. AWS (Route 53) or Designate (OpenStack).

In this proof-of-concept i've created a new source to create DNS records based on F5 Networks `VirtualServer` CRDs. There's two fields in the `VirtualServer` CRD that is of interest, the `host` and the `virtualServerAddress` fields.

To test this out i build `external-dns` locally and create a `kind` cluster, when i start `external-dns` locally it'll connect to the `kind` cluster and start to watch for `VirtalServer` CRDs.

One awesome feature in `external-dns` is that you can use the `inmemory` provider to store DNS records in memory, you don't need something live to connect to. Useful to do some manual testing of `external-dns`.

## Step-by-step
1. Clone my `external-dns` fork:
```
git clone https://github.com/mikejoh/external-dns.git 
``` 
2. Checkout my feature branch:
```
git checkout add-f5-virtualserver-source
```
3. Build `external-dns`:
```
make build
```
4. Create the `kind` cluster:
```
kind create cluster --config=external-dns-cluster.yaml
```
4. Install the CRDs shipped from F5 Networks (defined as part of the `k8s-bigip-ctlr` repository):
```
kubectl create -f https://raw.githubusercontent.com/F5Networks/k8s-bigip-ctlr/master/docs/config_examples/customResourceDefinitions/customresourcedefinitions.yml
```
5. Start `external-dns` locally:
```
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
6. Create a `VirtualServer` object:
```
kubectl create -f f5-virtual-server-example.yaml
```

You'll see `external-dns` create records in the `inmemory` provider by watching the standard output of the `external-dns` binary.

If you want to test the `crd` source, which watches for `DNSEndpoint` CRDs (provided by `external-dns`):
1. Install the CRD:
```
kubectl create -f https://raw.githubusercontent.com/kubernetes-sigs/external-dns/master/docs/contributing/crd-source/crd-manifest.yaml
```
2. Create a `DNSEndpoint` object:
```
kubectl create -f dnsendpoint-example.yaml
```

