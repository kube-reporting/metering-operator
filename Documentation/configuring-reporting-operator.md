# Configuring reporting-operator

The reporting-operator is responsible for collecting data from Prometheus, storing the metrics in Presto, running report queries against Presto, and exposing their results via an HTTP API.

Configuring the operator is primarily done within a `MeteringConfig` Custom Resources's `spec.reporting-operator.spec` section.

## Prometheus connection

When deploying Metering on an Openshift cluster, the reporting-operator defaults to using the Thanos Querier interface in the openshift-monitoring namespace: `https://thanos-querier.openshift-monitoring.svc:9091/`.

In the case where you aren't deploying Metering on an Openshift cluster, then you need to specify the Prometheus connection URL yourself in the `MeteringConfig` custom resource.

If you don't have a dedicated Prometheus instance available, we recommend following the instructions listed in the [kube-prometheus][kube-prometheus] project.

Below is an example of configuring Metering to use the service `prometheus` on port 9090 in the `cluster-monitoring` namespace:

```yaml
spec:
  reporting-operator:
    spec:
      config:
        prometheus:
          url: "http://prometheus.cluster-monitoring.svc:9090"
```

To secure the connection to Prometheus, the default Metering installation uses the Openshift certificate authority. If your Prometheus instance uses a different CA, the CA can be injected through a ConfigMap:

```yaml
spec:
  reporting-operator:
    spec:
      config:
        prometheus:
          certificateAuthority:
            useServiceAccountCA: false
            configMap:
              enabled: true
              create: true
              name: reporting-operator-certificate-authority-config
              filename: "internal-ca.crt"
              value: |
                -----BEGIN CERTIFICATE-----
                (snip)
                -----END CERTIFICATE-----
```

Alternatively, to use the system certificate authorities for publicly valid certificates, set both `useServiceAccountCA` and `configMap.enabled` to false.

Reporting-operator can also be configured to use a specified bearer token to auth with Prometheus:

```yaml
spec:
  reporting-operator:
    spec:
      config:
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

For Openshift, the metering operator exposes a [Route][route] by default, and for anything else you can use regular [Load Balancer][load-balancer-svc] or [Node Port][node-port-svc] [Kubernetes services][kube-svc].

### Openshift Route

Using an Openshift route has a few advantages over using a load balancer or node port service:

- Automatic DNS
- Automatic TLS based on the cluster CA

Additionally, on Openshift:

- We can take advantage of the [Openshift service serving certificates][service-certs] to protect the reporting API with TLS.
- We deploy the [Openshift OAuth proxy][oauth-proxy] as a side-car container for reporting-operator, which protects the reporting API with authentication.

There are a few ways to do authentication: you can use service account tokens for authentication, and/or you can also use a static username/password via an httpasswd file.
See the [openshift authentication](#openshift-authentication) section below for details on how authentication and authorization works.

### Load Balancer/Node Port services

While possible, using a LoadBalancer service or NodePort isn't currently recommended as the reporting-operator doesn't have any authentication methods available on non-openshift environments and exposing the API would result in your reporting being accessible to others.
This includes being able to download the raw collected data, reporting data, and the ability to push data as well.
If your NodePorts and/or LoadBalancers are not accessible to others, then you can consider enabling this, however it is still recommended to look into alternatives such as exposing metering using an Ingress controller that can provide authentication.

Exposing the reporting API is as simple as changing the type of `Service` used for the reporting-operator:

```yaml
apiVersion: metering.openshift.io/v1
kind: MeteringConfig
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

```yaml
kubectl -n $METERING_NAMESPACE get service reporting-operator -o wide
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP     PORT(S)          AGE   SELECTOR
reporting-operator   LoadBalancer   172.30.21.195   35.227.172.86   8080:32313/TCP   55m   app=reporting-operator
```

In this example the externalIP of the LoadBalancer is `35.227.172.86` and the port is 8080:

```bash
curl "http://35.227.172.86:8080/api/v1/reports/get?name=cluster-memory-capacity-hourly&namespace=openshift-metering&format=tab"
```

### Openshift Authentication

By default, the reporting API is secured with TLS and authentication. This is done by configuring the reporting-operator to deploy a pod containing both the reporting-operator's container, and a sidecar container running [Openshift auth-proxy](https://github.com/openshift/oauth-proxy).

If you want to manually configure authentication in Openshift, disable it entirely, or want more information on the reporting-operator options, see the [manual setup](#manual-setup) section.

In order to access the reporting API, the metering operator exposes a route. Once that route has been installed, you can run the command below to get the route's hostname.

```bash
METERING_ROUTE_HOSTNAME=$(oc -n $METERING_NAMESPACE get routes metering -o json | jq -r '.status.ingress[].host')
```

Also, make sure the `METERING_NAMESPACE` environment variable is set before continuing on with the next sections.

To authenticate and access the reporting API, you can do one of two options:

##### Authenticate using a service account token
See the [token authentication](#token-authentication) section for more information on how to extend the capabilities of this method using permissions.

With this method, we use the token in the reporting operator's service account, and pass that bearer token to the Authorization header in the following command:

```bash
TOKEN=$(oc -n $METERING_NAMESPACE serviceaccounts get-token reporting-operator)
curl -H "Authorization: Bearer $TOKEN" -k "https://$METERING_ROUTE_HOSTNAME/api/v1/reports/get?name=[Report Name]&namespace=$METERING_NAMESPACE&format=[Format]"
```

Be sure to replace the `name=[Report Name]` and `format=[Format]` parameters in the URL above.

##### Authenticate using a username and password
We are able to do basic authentication using a username and password combination, which is specified in the contents of a htpasswd file. We, by default, create a secret containing an empty htpasswd data. You can, however, configure the `reporting-operator.spec.authProxy.htpasswd.data` and `reporting-operator.spec.authProxy.htpasswd.createSecret: true` keys to use this method. See the [basic authentication](#basic-authentication-usernamepassword) section for more information.

Once you have specified the above in your `MeteringConfig` CR, you can run the following command:

```bash
curl -u testuser:password123 -k "https://$METERING_ROUTE_HOSTNAME/api/v1/reports/get?name=[Report Name]&namespace=$METERING_NAMESPACE&format=[Format]"
```

Be sure to replace `testuser:password123` with a valid username and password combination.

### Manual Setup

In order to manually configure, or disable OAuth in the reporting-operator, you need to set `spec.tls.enabled: false` in your `MeteringConfig` CR. **Warning:** this also disables all TLS/authentication between the reporting-operator, presto, and hive. You would need to manually configure these resources yourself.

Authentication can be enabled by configuring the options below.
Enabling authentication configures the reporting-operator pod to run the Openshift auth-proxy as a sidecar container in the pod.
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
For an example see the [advanced-auth.yaml][advanced-auth-config] example configuration.

Alternatively, you may use any role which has rules granting `get` permissions to `reports/export`.
Meaning: `get` access to the `export` _sub-resource_ of the `Report` resources in the namespace of the `reporting-operator`.
For example: `admin` and `cluster-admin`.

By default, the `reporting-operator` and `metering-operator` serviceAccounts both have these permissions, and their tokens may be used for authentication.
In this document, most examples will prefer this method.

#### Basic Authentication (username/password)

If `reporting-operator.spec.authProxy.htpasswd.data` is non-empty, its contents must be that of an [htpasswd file](https://httpd.apache.org/docs/2.4/programs/htpasswd.html).
When set, you can use [HTTP basic authentication][basic-auth-rfc] to provide your username and password that has a corresponding entry in the `htpasswd.data` contents.

Consider the following example as reference:

```yaml
apiVersion: metering.openshift.io/v1
kind: MeteringConfig
metadata:
  name: "operator-metering"
spec:
  reporting-operator:
    spec:
      authProxy:
        enabled: true
        
        # htpasswd.data can contain htpasswd file contents for allowing auth
        # using a static list of usernames and their password hashes.
        #
        # username is 'testuser' password is 'password123'
        # generated htpasswdData using: `htpasswd -nb -s testuser password123`
        # htpasswd:
        #   data: |
        #     testuser:{SHA}y/2sYAj5yrQIN4TL0YdPdmGNKpc=
        #
        # change REPLACEME to the output of your htpasswd command
        #
        # Set createSecret to TRUE which is mandatory if data is set
        htpasswd:
          createSecret: true
          data: |
            REPLACEME
```


[route]: https://docs.openshift.com/container-platform/3.11/dev_guide/routes.html
[kube-svc]: https://kubernetes.io/docs/concepts/services-networking/service/
[load-balancer-svc]: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
[node-port-svc]: https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
[service-certs]: https://docs.openshift.com/container-platform/3.11/dev_guide/secrets.html#service-serving-certificate-secrets
[oauth-proxy]: https://github.com/openshift/oauth-proxy
[expose-route-config]: ../manifests/metering-config/expose-route.yaml
[basic-auth-rfc]: https://tools.ietf.org/html/rfc7617
[advanced-auth-config]: ../manifests/metering-config/advanced-auth.yaml
[kube-prometheus]: https://github.com/coreos/kube-prometheus
