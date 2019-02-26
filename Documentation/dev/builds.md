# Overview

There are a number of builds and container images which are part of the Metering default "stack".
Each of them is built slightly differently.

- reporting-operator
  - Written in Go
  - Source is primarily in `pkg/`, and `cmd/` of this repo.
  - origin
    - source is https://github.com/operator-framework/operator-metering
    - Docker***REMOVED***le is `Docker***REMOVED***le.reporting-operator`
    - Docker image is [quay.io/coreos/metering-reporting-operator](https://quay.io/repository/coreos/metering-reporting-operator)
    - Built by this repos Jenkins CI.
  - OCP
    - Source: http://pkgs.devel.redhat.com/cgit/containers/ose-metering-reporting-operator/
    - OCP Docker***REMOVED***le is `Docker***REMOVED***le.reporting-operator.rhel`
    - OCP Docker image is [registry.access.redhat.com/openshift4/ose-metering-reporting-operator](https://registry.access.redhat.com/openshift4/ose-metering-reporting-operator)
    - Built in brew by OSBS
    - brew package name: `ose-metering-reporting-operator-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70735
- metering-helm-operator
  - Based on upstream helm repo fork
  - Written using mostly bash, jq, python and helm charts.
  - Source is primarily in `charts/`, `images/helm-operator/`
  - origin
    - Source: https://github.com/operator-framework/operator-metering
    - Docker***REMOVED***le is `Docker***REMOVED***le.metering-operator`
    - Docker image is [quay.io/coreos/metering-helm-operator](https://quay.io/repository/coreos/metering-helm-operator)
    - Built by this repos Jenkins CI.
  - OCP
    - Source: http://pkgs.devel.redhat.com/cgit/containers/ose-metering-helm-operator/
    - OCP Docker***REMOVED***le is `Docker***REMOVED***le.metering-operator.rhel`
    - OCP Docker image is [registry.access.redhat.com/openshift4/ose-metering-helm-operator](https://registry.access.redhat.com/openshift4/ose-metering-helm-operator)
    - Built in brew by OSBS
    - brew package name: `ose-metering-helm-operator-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70824
- helm
  - Written in Go
  - origin
    - Source is https://github.com/operator-framework/helm/blob/metering-v2.8.2
    - Docker***REMOVED***le is https://github.com/operator-framework/helm/blob/metering-v2.8.2/Docker***REMOVED***le
    - Docker image is [quay.io/coreos/helm](https://quay.io/repository/coreos/helm)
    - Built by quay.io
  - OCP
    - Source is  http://pkgs.devel.redhat.com/cgit/containers/ose-metering-helm/
    - OCP Docker***REMOVED***le is https://github.com/operator-framework/helm/blob/release-4.0/Docker***REMOVED***le.rhel
    - OCP Docker image is [registry.access.redhat.com/openshift4/ose-metering-helm](https://registry.access.redhat.com/openshift4/ose-metering-helm)
    - Built in brew by OSBS
    - brew package name: `ose-metering-helm-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70736
- presto
  - Written in Java, uses [maven][maven] as a project build tool.
  - origin
    - Source is https://github.com/operator-framework/presto/tree/metering-0.212
    - Docker***REMOVED***le is https://github.com/operator-framework/presto/blob/metering-0.212/Docker***REMOVED***le
    - Docker image is [quay.io/coreos/presto](https://quay.io/repository/coreos/presto)
    - Built by quay.io
  - OCP
    - Source is http://pkgs.devel.redhat.com/cgit/containers/presto/
    - OCP Docker***REMOVED***le is https://github.com/operator-framework/presto/blob/release-4.0/Docker***REMOVED***le.rhel
    - OCP Docker image is [registry.access.redhat.com/openshift4/ose-presto](https://registry.access.redhat.com/openshift4/ose-presto)
    - Built in brew by OSBS
    - brew package name: `presto-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70772
- hive
  - Written in Java, uses [maven][maven] as a project build tool.
  - origin
    - Source is https://github.com/operator-framework/hive/tree/metering-2.3.3
    - Docker***REMOVED***le is https://github.com/operator-framework/hive/tree/metering-2.3.3/Docker***REMOVED***le
    - Docker image is [quay.io/coreos/hive](https://quay.io/repository/coreos/hive)
    - Built by quay.io
  - OCP
    - Source is http://pkgs.devel.redhat.com/cgit/containers/hive/
    - OCP Docker***REMOVED***le is https://github.com/operator-framework/hive/tree/release-4.0/Docker***REMOVED***le.rhel
    - OCP Docker image is [registry.access.redhat.com/openshift4/ose-hive](https://registry.access.redhat.com/openshift4/ose-hive)
    - Built in brew by OSBS
    - brew package name: `hive-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70894
- hadoop
  - Written in Java, uses [maven][maven] as a project build tool.
  - origin
    - Source is https://github.com/operator-framework/hadoop/tree/metering-3.1.1
    - Docker***REMOVED***le is https://github.com/operator-framework/hadoop/tree/metering-3.1.1/Docker***REMOVED***le
    - Docker image is [quay.io/coreos/hadoop](https://quay.io/repository/coreos/hadoop)
    - Built by quay.io
  - OCP
    - Source is http://pkgs.devel.redhat.com/cgit/containers/hadoop/
    - OCP Docker***REMOVED***le is https://github.com/operator-framework/hadoop/tree/release-4.0/Docker***REMOVED***le.rhel
    - OCP Docker image is [registry.access.redhat.com/openshift4/ose-hadoop](https://registry.access.redhat.com/openshift4/ose-hadoop)
    - Built in brew by OSBS
    - brew package name: `hadoop-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70822

# OCP builds

OCP builds are done by a combination of systems within the Red Hat internal network.
Many links below will require you to be on the Red Hat network in some form (the VPN is a good option).

The major components are:
- [Openshift Build System (OSBS)][osbs]
  - This builds docker images in brew/pulp in an isolated build environment.
  - OSBS builds for OCP are managed by Automated Release Team (ART).
  - See https://mojo.redhat.com/docs/DOC-1179058 for the document managed by
    ART on OCP automated releases.
- [Project New Castle (pnc)][pnc]
  - Builds Java projects that use [maven][maven].
- [Brew (koji)][brew]
  - Builds, Images, rpms, java artifacts, etc all get pushed here.
- [dist-git][dist-git]
  - Holds copies of our repos that are synced and managed by the Automated Release Team (ART).
  - Contains modi***REMOVED***ed Docker***REMOVED***les: They use a tool [OIT][oit] that adds some additional changes to the repo before syncing it to dist-git.
  - One dist-git repo per docker image. Each repo is for one Docker***REMOVED***le + other ***REMOVED***les used by OSBS.

For reporting-operator, helm, and metering-helm-operator, they all are written primarily in Go, and have any dependencies vendored, so OSBS can build them directly without any additional requirements once ART runs [oit][oit] to sync the repositories and update the Docker***REMOVED***les.

For Presto, Hive, and Hadoop, these are all written in Java, and their dependencies are fetched using [maven][maven], meaning extra work is required to build these, since OSBS doesn't allow downloading things outside the network.
To handle this we use [PNC][pnc] to perform builds using maven and push the artifacts to brew.

## Project New Castle (PNC)

Project New Castle (PNC) is basically an isolated build environment that proxies maven downloads ensuring downloads come from an internal maven repo.
It also proxies from maven central anything that isn't available yet, and uses persistent caching to ensure that once something is downloaded it doesn't change.
Once it's done the builds, we do a push to brew.

The Openshift Metering "product" is found at: http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/product/50

Current product versions:

- Openshift Metering 0.7
  - Brew Tag Pre***REMOVED***x: `openshift-metering-0.7-pnc`

We currently have 4 projects in PNC:

- presto
  - http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/207
  - build con***REMOVED***gs:
    - presto-0.212: http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/build-con***REMOVED***gs/571
- hive
  - http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/220
  - build con***REMOVED***gs:
    - hive-2.3.3: http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/220/build-con***REMOVED***gs/636
- hadoop
  - http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/219
  - build con***REMOVED***gs:
    - hadoop-3.1.1: http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/219/build-con***REMOVED***gs/633
- prometheus-jmx-exporter
  - http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/208
  - build con***REMOVED***gs:
    - prometheus-jmx-exporter-0.3.1: http://orch.cloud.pnc.engineering.redhat.com/pnc-web/#/projects/208/build-con***REMOVED***gs/578

### PNC Brew Push

Read the PNC documentation on closing milestones: https://docs.engineering.redhat.com/pages/viewpage.action?pageId=44534467

Also read https://docs.engineering.redhat.com/display/JP/Integration+with+Brew for details on the brew tags that need to be created prior to the push.
Usually having tags added is as simple as ***REMOVED***ling a Jira ticket in the RCM project (Ex: https://projects.engineering.redhat.com/browse/RCM-43044).

[osbs]: https://osbs.readthedocs.io/en/latest/
[pnc]: https://docs.engineering.redhat.com/display/JP/User%27s+guide
[brew]: https://brewweb.engineering.redhat.com/brew/
[dist-git]: http://pkgs.devel.redhat.com/cgit/
[oit]: https://github.com/openshift/enterprise-images/blob/d44b833f8696102364f2526eaf130a961eb4cf56/oit.py
[maven]: https://maven.apache.org/
[pnc-brew]: https://docs.engineering.redhat.com/display/JP/Integration+with+Brew
