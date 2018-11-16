## Prometheus URL

By default, the Metering assumes that your Prometheus service is available at `http://prometheus-k8s.monitoring.svc:9090` within the cluster.
If your not using [kube-prometheus][kube-prometheus], then you will need to override the `reporting-operator.con***REMOVED***g.prometheusURL` con***REMOVED***guration option.

Below is an example of con***REMOVED***guring Metering to use the service `prometheus` on port 9090 in the `cluster-monitoring` namespace:

```
spec:
  reporting-operator:
    spec:
      con***REMOVED***g:
        prometheusURL: "http://prometheus.cluster-monitoring.svc:9090"
```

> Note: currently we do not support https connections or authentication to Prometheus except for in Openshift, but support for it is being developed.

## Resource requests and limits

You can adjust the cpu, memory, or storage resource requests and/or limits for pods and volumes.
See [resource-limits.yaml][resource-limits] for an example of setting resource request and limits for each component.

## Node Selectors

If you want to run the metering components on speci***REMOVED***c sets of nodes then you can set nodeSelectors on each component to control where each component of metering is scheduled to.
See [node-selectors.yaml][node-selectors-con***REMOVED***g] for an example of setting node selectors for each component.

## Image repositories and tags

You can override the image repositories and versions to test pre-releases or to deploy an image built by our CI for PRs or testing.
See [latest-versions.yaml][latest-versions] for an example of setting the repository and image tag for each component to use.

[latest-versions]: ../manifests/metering-con***REMOVED***g/latest-versions.yaml
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
[node-selectors-con***REMOVED***g]: ../manifests/metering-con***REMOVED***g/custom-node-selectors.yaml
[resource-limits]: ../manifests/metering-con***REMOVED***g/resource-limits.yaml
