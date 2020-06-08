# Current Release: 4.6

## Release Notes

- [#1214](https://github.com/kube-reporting/metering-operator/pull/1214) Add initial support for configuring the Hive Metastore database to reference a secret containing the base64 encrypted username and password credentials.
- [#1224](https://github.com/kube-reporting/metering-operator/pull/1224) Improve performance of the metering-ansible-operator by "finalizing" the `meteringconfig_spec_overrides` dictionary.
- [#1226](https://github.com/kube-reporting/metering-operator/pull/1226) Re-generate assets and YAML manifests to point to the 4.6 images.

### Bug Fixes

- [#1228](https://github.cohttps://github.com/kube-reporting/metering-operator/pull/1228m/kube-reporting/metering-operator/pull/1228) Reference the correct serviceaccount to allow Prometheus to scrap Metering endpoints properly.
- [#1229](https://github.com/kube-reporting/metering-operator/pull/1229) Migrate to obtaining the service-serving CA bundle from a ConfigMap.
- [#1235](https://github.com/kube-reporting/metering-operator/pull/1235) Update charts to expose the metering-ansible-operator metrics properly.
