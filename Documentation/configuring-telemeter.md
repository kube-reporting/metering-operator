# Configuring Metering for use with Telemeter

[Telemeter](https://github.com/openshift/telemeter) is a service that implements a Prometheus federation service. Metrics can be collected across multiple clusters and worked with to do reporting of cluster resource availability, usage, etc. Metering comes with several built-in queries for consuming Telemeter data, but it is disabled by default and must be configured before use.

## Prometheus connection

By default, Metering will use the in-cluster Openshift Prometheus instance. It is necessary to change the certificate authority configuration and Prometheus URL to point to the instance of Telemeter you wish to use. For detailed instructions, see [Prometheus connection](configuring-reporting-operator.md#prometheus-url); however, it is likely that at least the authentication/certificate authority section and the Prometheus URL will need to be set.

## Installing Telemeter queries

Metering comes with queries built around working with Openshift cluster metrics. However, for use with Telemeter, these should be disabled and the optional Telemeter queries enabled:

```yaml
spec:
  # disable Openshift metric queries and reports
  openshift-reporting:
    enabled: false
  # enable Telemeter metric queries and reports
  telemeter-reporting:
    enabled: true
```

## Metering CR

A complete Metering configuration resource with options set for connecting to a Telemeter Prometheus instance and using Telemeter queries will look somewhat as follows:

```yaml
# metering-telemeter.yaml
apiVersion: metering.openshift.io/v1alpha1
kind: Metering
metadata:
  name: operator-metering
spec:
  # disable Openshift metric queries and reports
  openshift-reporting:
    enabled: false
  # enable Telemeter metric queries and reports
  telemeter-reporting:
    enabled: true
  # configure connection to Telemeter Prometheus
  reporting-operator:
    spec:
      config:
        prometheusURL: "https://telemeter.api.mycompany.com"
        prometheusCAFile: "/path/to/my/ca.crt" # or "" to validate against system certificate authorities
        prometheusImporter:
          auth:
            # don't use in-cluster token authentication; instead use provided token
            useServiceAccountToken: false
            tokenSecret:
              enabled: true
              create: true
              value: "abc-123"
```