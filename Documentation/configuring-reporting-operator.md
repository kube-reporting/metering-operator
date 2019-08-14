# Con***REMOVED***guring reporting-operator

The reporting-operator is responsible for collecting data from Prometheus, storing the metrics in Presto, running report queries against Presto, and exposing their results via an HTTP API.
Con***REMOVED***guring the operator is primarily done within a `MeteringCon***REMOVED***g` Custom Resources's `spec.reporting-operator.spec` section.

## Prometheus connection

Depending on how you installed Metering, the default Prometheus URL varies.
If you installed for Openshift then the default assumes Prometheus is available at `https://prometheus-k8s.openshift-monitoring.svc:9091/`.
Otherwise it assumes that your Prometheus service is available at `http://prometheus-k8s.monitoring.svc:9090`.
If you're not on Openshift and aren't using [kube-prometheus][kube-prometheus], then you will need to override the `reporting-operator.con***REMOVED***g.prometheus.url` con***REMOVED***guration option.

Below is an example of con***REMOVED***guring Metering to use the service `prometheus` on port 9090 in the `cluster-monitoring` namespace:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        prometheus:
          url: "http://prometheus.cluster-monitoring.svc:9090"
```

To secure the connection to Prometheus, the default Metering installation uses the Openshift certi***REMOVED***cate authority. If your Prometheus instance uses a different CA, the CA can be injected through a Con***REMOVED***gMap:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        prometheus:
          certi***REMOVED***cateAuthority:
            useServiceAccountCA: false
            con***REMOVED***gMap:
              enabled: true
              create: true
              name: reporting-operator-certi***REMOVED***cate-authority-con***REMOVED***g
              ***REMOVED***lename: "internal-ca.crt"
              value: |
                -----BEGIN CERTIFICATE-----
                (snip)
                -----END CERTIFICATE-----
```

Alternatively, to use the system certi***REMOVED***cate authorities for publicly valid certi***REMOVED***cates, set both `useServiceAccountCA` and `con***REMOVED***gMap.enabled` to false.

Reporting-operator can also be con***REMOVED***gured to use a speci***REMOVED***ed bearer token to auth with Prometheus:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        prometheus:
          metricsImporter:
            auth:
              useServiceAccountToken: false
              tokenSecret:
                enabled: true
                create: true
                value: "abc-123"
```

## Exposing the reporting API

There are two ways to expose the reporting API depending on if you're using regular Kubernetes, or Openshift.

For Openshift, the metering operator exposes a [Route][route] by default, and for anything ***REMOVED*** you can use regular [Load Balancer][load-balancer-svc] or [Node Port][node-port-svc] [Kubernetes services][kube-svc].

### Openshift Route

Using an Openshift route has a few advantages over using a load balancer or node port service:

- Automatic DNS
- Automatic TLS based on the cluster CA

Additionally, on Openshift:

- We can take advantage of the [Openshift service serving certi***REMOVED***cates][service-certs] to protect the reporting API with TLS.
- We deploy the [Openshift OAuth proxy][oauth-proxy] as a side-car container for reporting-operator, which protects the reporting API with authentication.

There are a few ways to do authentication: you can use service account tokens for authentication, and/or you can also use a static username/password via an httpasswd ***REMOVED***le.
See the [openshift authentication](#openshift-authentication) section below for details on how authentication and authorization works.

### Load Balancer/Node Port services

While possible, using a LoadBalancer service or NodePort isn't currently recommended as the reporting-operator doesn't have any authentication methods available on non-openshift environments and exposing the API would result in your reporting being accessible to others.
This includes being able to download the raw collected data, reporting data, and the ability to push data as well.
If your NodePorts and/or LoadBalancers are not accessible to others, then you can consider enabling this, however it is still recommended to look into alternatives such as exposing metering using an Ingress controller that can provide authentication.

Exposing the reporting API is as simple as changing the type of `Service` used for the reporting-operator:

```
apiVersion: metering.openshift.io/v1
kind: MeteringCon***REMOVED***g
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

By default, the reporting API is secured with TLS and authentication. This is done by con***REMOVED***guring the reporting-operator to deploy a pod containing both the reporting-operator's container, and a sidecar container running [Openshift auth-proxy](https://github.com/openshift/oauth-proxy).

If you want to manually con***REMOVED***gure authentication in Openshift, disable it entirely, or want more information on the reporting-operator options, see the [manual setup](#manual-setup) section.

In order to access the reporting API, the metering operator exposes a route. Once that route has been installed, you can run the command below to get the route's hostname.

```
METERING_ROUTE_HOSTNAME=$(oc -n $METERING_NAMESPACE get routes metering -o json | jq -r '.status.ingress[].host')
```

Also, make sure the `METERING_NAMESPACE` environment variable is set before continuing on with the next sections.

To authenticate and access the reporting API, you can do one of two options:

##### Authenticate using a service account token
See the [token authentication](#token-authentication) section for more information on how to extend the capabilities of this method using permissions.

With this method, we use the token in the reporting operator's service account, and pass that bearer token to the Authorization header in the following command:

```
TOKEN=$(oc -n $METERING_NAMESPACE serviceaccounts get-token reporting-operator)
curl -H "Authorization: Bearer $TOKEN" -k "https://$METERING_ROUTE_HOSTNAME/api/v1/reports/get?name=[Report Name]&namespace=$METERING_NAMESPACE&format=[Format]"
```

Be sure to replace the `name=[Report Name]` and `format=[Format]` parameters in the URL above.

##### Authenticate using a username and password
We are able to do basic authentication using a username and password combination, which is speci***REMOVED***ed in the contents of a htpasswd ***REMOVED***le. We, by default, create a secret containing an empty htpasswd data. You can, however, con***REMOVED***gure the `reporting-operator.spec.authProxy.htpasswd.data` and `reporting-operator.spec.authProxy.htpasswd.createSecret` keys to use this method. See the [basic authentication](#basic-authentication-usernamepassword) section for more information.

Once you have speci***REMOVED***ed the above in your `MeteringCon***REMOVED***g` CR, you can run the following command:
```
curl -u testuser:password123 -k "https://$METERING_ROUTE_HOSTNAME/api/v1/reports/get?name=[Report Name]&namespace=$METERING_NAMESPACE&format=[Format]"
```

Be sure to replace `testuser:password123` with a valid username and password combination.

### Manual Setup

In order to manually con***REMOVED***gure, or disable OAuth in the reporting-operator, you need to set `spec.tls.enabled: false` in your `MeteringCon***REMOVED***g` CR. Warning: this also disables all TLS/authentication between the reporting-operator, presto, and hive. You would need to manually con***REMOVED***gure these resources yourself.

Authentication can be enabled by con***REMOVED***guring the options below.
Enabling authentication con***REMOVED***gures the reporting-operator pod to run the Openshift auth-proxy as a sidecar container in the pod.
This adjusts the ports so that the reporting-operator API isn't exposed directly, but instead is proxied to via the auth-proxy sidecar container.

- `reporting-operator.spec.authProxy.enabled`
- `reporting-operator.spec.authProxy.cookie.createSecret`
- `reporting-operator.spec.authProxy.cookie.seed`

You need to set `reporting-operator.spec.authProxy.enabled` and `reporting-operator.spec.authProxy.cookie.createSecret` to true and `reporting-operator.spec.authProxy.cookie.seed` to a 32-character random string.

You can generate a 32-character random string using the command `$ openssl rand -base64 32 | head -c32; echo`.

#### Token Authentication

When the following options are set to true, authentication using a bearer token is enabled for the reporting REST API.
Bearer tokens may come from serviceAccounts or users.

- `reporting-operator.spec.authProxy.subjectAccessReview.enabled`
- `reporting-operator.spec.authProxy.delegateURLs.enabled`

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

If `reporting-operator.spec.authProxy.htpasswd.data` is non-empty, its contents must be that of an [htpasswd ***REMOVED***le](https://httpd.apache.org/docs/2.4/programs/htpasswd.html).
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
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
