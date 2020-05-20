# Configuring Real DNS and auto-TLS

By default, `mink` runs a `default-domain` `Job` that attempts to set up usable
URLs out-of-the-box with `xip.io`. However, this is unsuitable for production
workloads and doesn't work with auto-TLS.

To setup proper DNS, start by identifying the network endpoint for external
services:

```shell
$ kubectl get svc -nmink-system envoy-external

NAME             TYPE           CLUSTER-IP   EXTERNAL-IP     PORT(S)                      AGE
envoy-external   LoadBalancer   10.0.13.82   34.94.155.233   80:32533/TCP,443:30429/TCP   5d19h
```

**If you got back an actual IP address, then you should set up an `A` record for
`*` with this IP.**

On AWS, this will look like:

```shell
$ kubectl get svc -nmink-system envoy-external

NAME             TYPE           CLUSTER-IP      EXTERNAL-IP                                                               PORT(S)                      AGE
envoy-external   LoadBalancer   10.100.172.92   ad42e899328a046578b889efe9b555e4-1779865644.us-west-2.elb.amazonaws.com   80:30417/TCP,443:31728/TCP   62d
```

**If you got back a hostname, then you should set up a `CNAME` record for `*`
with this hostname.**

Last, tell us to use this domain with:

```shell
kubectl patch -nmink-system configmap/config-domain \
  --type='json' \
  --patch='[{"op": "replace", "path": "/data", "value":{"your-domain.dev": ""}}]'
```

Once real DNS is configured, `mink` will migrate all of your services to the new
domain, and use ACME HTTP01 challenges to provision certificates.
