# Common configuration

This document contains example configurations for configuration that spans one or more components.

## Resource requests and limits

You can adjust the cpu, memory, or storage resource requests and/or limits for pods and volumes.
See [resource-limits.yaml][resource-limits] for an example of setting resource request and limits for each component.

## Node Selectors

If you want to run the metering components on specific sets of nodes then you can set nodeSelectors on each component to control where each component of metering is scheduled to.
See [node-selectors.yaml][node-selectors-config] for an example of setting node selectors for each component.

## Image repositories and tags

You can override the image repositories and versions to test pre-releases or to deploy an image built by our CI for PRs or testing.
See [latest-versions.yaml][latest-versions] for an example of setting the repository and image tag for each component to use.

[latest-versions]: ../manifests/metering-config/latest-versions.yaml
[kube-prometheus]: https://github.com/coreos/prometheus-operator/tree/master/contrib/kube-prometheus
[node-selectors-config]: ../manifests/metering-config/custom-node-selectors.yaml
[resource-limits]: ../manifests/metering-config/resource-limits.yaml
[route]: https://docs.openshift.com/container-platform/3.11/dev_guide/routes.html
[kube-svc]: https://kubernetes.io/docs/concepts/services-networking/service/
[load-balancer-svc]: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
[node-port-svc]: https://kubernetes.io/docs/concepts/services-networking/service/#nodeport
[service-certs]: https://docs.openshift.com/container-platform/3.11/dev_guide/secrets.html#service-serving-certificate-secrets
[oauth-proxy]: https://github.com/openshift/oauth-proxy
[expose-route-config]: ../manifests/metering-config/expose-route.yaml
