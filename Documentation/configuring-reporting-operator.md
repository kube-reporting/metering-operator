# Con***REMOVED***guring reporting-operator

reporting-operator is responsible for collecting data from Prometheus, storing the metrics in Presto, running report queries against Presto, and exposing their results via an HTTP API.
Con***REMOVED***guring the operator is done primarily within a `Metering` CR's `spec.reporting-operator.spec` section.

## Prometheus URL

Depending on how you installed Metering, the default Prometheus URL varies.
If you installed for Openshift then the default assumes Prometheus is available at `https://prometheus-k8s.openshift-monitoring.svc:9091/`.
Otherwise it assumes that your Prometheus service is available at `http://prometheus-k8s.monitoring.svc:9090`.
If your not Openshift or using [kube-prometheus][kube-prometheus] on non-openshift clusters, then you will need to override the `reporting-operator.con***REMOVED***g.prometheusURL` con***REMOVED***guration option.

Below is an example of con***REMOVED***guring Metering to use the service `prometheus` on port 9090 in the `cluster-monitoring` namespace:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        prometheusURL: "http://prometheus.cluster-monitoring.svc:9090"
```

> Note: currently we do not support https connections or authentication to Prometheus except for in Openshift, but support for it is being developed.

## Exposing the reporting API

There are two ways to expose the reporting API depending on if you're using regular Kubernetes, or Openshift.

For Openshift, you can expose a [Route][route], and for anything ***REMOVED*** you can use regular [Load Balancer][load-balancer-svc] or [Node Port][node-port-svc] [Kubernetes services][kube-svc].

### Openshift Route

Using an Openshift route has a few advantages over using a load balancer or node port service:

- Automatic DNS
- Automatic TLS based on the cluster CA

Additionally, on Openshift:

- We can take advantage of the [Openshift service serving certi***REMOVED***cates][service-certs] to protect the reporting API with TLS
- We deploy the [Openshift OAuth proxy][oauth-proxy] as a side-car container for reporting-operator, which protects the reporting API with authentication

There are a few ways to do authentication, you can use service account tokens for authentication, and/or you can also use a static username/password via an httpasswd ***REMOVED***le.
See the [openshift authentication](#openshift-authentication) section below for details on how authentication and authorization works.
See the [expose-route.yaml][expose-route-con***REMOVED***g] con***REMOVED***guration for an example of setting enabling an Openshift route and con***REMOVED***guring authentication with both options enabled.

Make sure you modify the `reporting-operator.spec.authProxy.httpasswdData` and `reporting-operator.spec.authProxy.cookieSeed` values.

Once installed with the customized con***REMOVED***guration to enable the route, you should query in your namespace to check for the route:

```
oc -n openshift-metering get routes
NAME       HOST/PORT                                         PATH      SERVICES             PORT      TERMINATION   WILDCARD
metering   metering-openshift-metering.apps.example.com                reporting-operator   http      reencrypt     None
```

This will provide a URL you can use to access the reporting API.

To authenticate, you can do one of two options depending on which authentication methods you enabled:

To authenticate with a service account, pass it using the Authorization header as a bearer token:

```
TOKEN=$(oc -n openshift-metering serviceaccounts get-token reporting-operator)
curl -H "Authorization: Bearer $TOKEN" -k "https://metering-openshift-metering.apps.example.com/api/v1/reports/get?name=cluster-memory-capacity-hourly&namespace=openshift-metering&format=tab"
```

And to authenticate using a username and password, use basic authentication:

```
curl -u testuser:password123 -k "https://metering-openshift-metering.apps.example.com/api/v1/reports/get?name=cluster-memory-capacity-hourly&namespace=openshift-metering&format=tab"
```

### Load Balancer/Node Port services

Using a LoadBalancer service or NodePort while possible, isn't currently recommended as the reporting-operator doesn't have any authentication methods available on non-openshift environments and exposing the API would result in your reporting being accessible to others.
This includes being able to download the raw collected data, reporting data, and the ability to push data as well.
If your NodePorts and/or LoadBalancers are not accessible to others, then you can consider enabling this, however it is still recommended to look into alternatives such as exposing metering using an Ingress controller that can provide authentication.

Exposing the reporting API is as simple as changing the type of `Service` used for the reporting-operator:

```
apiVersion: metering.openshift.io/v1alpha1
kind: Metering
metadata:
  name: "operator-metering"
spec:
  reporting-operator:
    spec:
      service:
        type: LoadBalancer
        # Can also be:
        # type: NodePort
        # you can also set the nodePort directly if one hasn't been set previously:
        # nodePort: 32313
```

Accessing it is dependent on what kind of service you created, but information on the LoadBalancer or NodePort can be found using kubectl:

```
kubectl -n $METERING_NAMESPACE get service reporting-operator -o wide
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)          AGE   SELECTOR
reporting-operator   LoadBalancer   172.30.21.195   35.227.172.86   8080:32313/TCP   55m   app=reporting-operator
```

In this example the externalIP of the LoadBalancer is `35.227.172.86` and the port is 8080:

```
curl "http://35.227.172.86:8080/api/v1/reports/get?name=cluster-memory-capacity-hourly&namespace=openshift-metering&format=tab"
```

### Openshift Authentication

Authentication can be enabled by setting the options below to true.
Enabling authentication con***REMOVED***gures the reporting-operator pod to run the Openshift auth-proxy as a sidecar container in the pod.
This adjusts the ports so that the reporting-operator API isn't exposed directly, but instead is proxied to via the auth-proxy sidecar container.

- `reporting-operator.spec.authProxy.enabled`
- `reporting-operator.spec.authProxy.cookieSeed`

#### Token Authentication

When the following options are set to true, authentication using a bearer token is enabled for the reporting REST API.
Bearer tokens may come from serviceAccounts or users.

- `reporting-operator.spec.authProxy.subjectAccessReviewEnabled`
- `reporting-operator.spec.authProxy.delegateURLsEnabled`

When authentication is enabled, the Bearer token used to query the reporting API of the user or serviceAccount must be granted access using one of the following roles:

- `report-exporter`
- `reporting-admin`
- `reporting-viewer`
- `metering-admin`
- `metering-viewer`

The metering-operator is capable of creating RoleBindings for you, granting these permissions by specifying a list of subjects in the Metering `spec.permissions` section.
For an example see the [advanced-auth.yaml][advanced-auth-con***REMOVED***g] example con***REMOVED***guration.

Alternatively, you may use any role which has rules granting `get` permissions to `reports/export`.
Meaning: `get` access to the `export` _sub-resource_ of the `Report` resources in the namespace of the `reporting-operator`.
For example: `admin` and `cluster-admin`.

By default, the `reporting-operator` and `metering-operator` serviceAccounts both have these permissions, and their tokens may be used for authentication.
In this document, most examples will prefer this method.

#### Basic Authentication (username/password)

If `reporting-operator.spec.authProxy.htpasswdData` is non-empty, it's contents must be that of an [htpasswd ***REMOVED***le](https://httpd.apache.org/docs/2.4/programs/htpasswd.html).
When set, you can use [HTTP basic authentication][basic-auth-rfc] to provide your username and password that has a corresponding entry in the `htpasswdData` contents.

[route]: https://docs.openshift.com/container-platform/3.11/dev_guide/routes.html
[kube-svc]: https://kubernetes.io/docs/concepts/services-networking/service/
[load-balancer-svc]: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
[node-port-svc]: https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
[service-certs]: https://docs.openshift.com/container-platform/3.11/dev_guide/secrets.html#service-serving-certi***REMOVED***cate-secrets
[oauth-proxy]: https://github.com/openshift/oauth-proxy
[expose-route-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/expose-route.yaml
[basic-auth-rfc]: https://tools.ietf.org/html/rfc7617
[advanced-auth-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/advanced-auth.yaml
