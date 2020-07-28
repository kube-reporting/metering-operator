# Current Release: 4.6

## Release Notes

- [#1214](https://github.com/kube-reporting/metering-operator/pull/1214) Add initial support for configuring the Hive Metastore database to reference a secret containing the base64 encrypted username and password credentials.
- [#1224](https://github.com/kube-reporting/metering-operator/pull/1224) Improve performance of the metering-ansible-operator by "finalizing" the `meteringconfig_spec_overrides` dictionary.
- [#1226](https://github.com/kube-reporting/metering-operator/pull/1226) Re-generate assets and YAML manifests to point to the 4.6 images.
- [#1245](https://github.com/kube-reporting/metering-operator/pull/1245) Consolidate Hive Metastore configurable fields. You no longer need to specify `spec.hive.spec.metastore.storage.create: false` when using a non-default database for Hive Metastore.
- [#1253](https://github.com/kube-reporting/metering-operator/pull/1253) Bumped the Metering-related CRD versioning from v1beta1 to v1. The minimum kubernetes version in the Metering CSV manifest is now set to 1.18.
- [#1264](https://github.com/kube-reporting/metering-operator/pull/1264) Bumped the Kubernetes-related dependencies to 1.18.x.
- [#1330](https://github.com/kube-reporting/metering-operator/pull/1330) Updated the metering-ansible-operator to use the operator-sdk 0.19.x version. As part of that process, we removed the Ansible sidecar container and the Ansible logs are now viewable in the `operator` container. For more information, see the removal notice in the [0.18.0 migration guide](https://sdk.operatorframework.io/docs/migration/v0.18.0/#remove-the-file-binao-logs-for-ansible-based-operators).

### Bug Fixes

- [#1228](https://github.cohttps://github.com/kube-reporting/metering-operator/pull/1228m/kube-reporting/metering-operator/pull/1228) Reference the correct serviceaccount to allow Prometheus to scrap Metering endpoints properly.
- [#1229](https://github.com/kube-reporting/metering-operator/pull/1229) Migrate to obtaining the service-serving CA bundle from a ConfigMap.
- [#1235](https://github.com/kube-reporting/metering-operator/pull/1235) Update charts to expose the metering-ansible-operator metrics properly.
