# Overview

There are a number of builds and container images which are part of the Metering default "stack".
Each of them is built slightly differently.

- reporting-operator
  - Written in Go
  - Source is primarily in `pkg/`, and `cmd/` of this repo.
  - origin
    - source is https://github.com/operator-framework/operator-metering
    - Dockerfile is `Dockerfile.reporting-operator`
    - Docker image is [quay.io/openshift/origin-metering-reporting-operator](https://quay.io/repository/openshift/origin-metering-reporting-operator)
    - Built by origin CI using prow/ci-operator
  - OCP
    - Source: http://pkgs.devel.redhat.com/cgit/containers/ose-metering-reporting-operator/
    - OCP Dockerfile is `Dockerfile.reporting-operator.rhel`
    - OCP Docker image is TBD
    - Built in brew by OSBS
    - brew package name: `ose-metering-reporting-operator-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70735
- metering-helm-operator
  - Based on upstream helm repo fork
  - Written using mostly bash, jq, python and helm charts.
  - Source is primarily in `charts/`, `images/helm-operator/`
  - origin
    - Source: https://github.com/operator-framework/operator-metering
    - Dockerfile is `Dockerfile.metering-operator`
    - Docker image is [quay.io/openshift/origin-metering-helm-operator](https://quay.io/repository/openshift/origin-metering-helm-operator)
    - Built by origin CI using prow/ci-operator
  - OCP
    - Source: http://pkgs.devel.redhat.com/cgit/containers/ose-metering-helm-operator/
    - OCP Dockerfile is `Dockerfile.metering-operator.rhel`
    - OCP Docker image is TBD
    - Built in brew by OSBS
    - brew package name: `ose-metering-helm-operator-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70824
- helm
  - Written in Go
  - origin
    - Source is https://github.com/operator-framework/helm/blob/master
    - Dockerfile is https://github.com/operator-framework/helm/blob/master/Dockerfile
    - Docker image is [quay.io/openshift/origin-metering-helm](https://quay.io/repository/openshift/origin-metering-helm)
    - Build by origin CI using prow/ci-operator
  - OCP
    - Source is  http://pkgs.devel.redhat.com/cgit/containers/ose-metering-helm/
    - OCP Dockerfile is https://github.com/operator-framework/helm/blob/master/Dockerfile.rhel
    - OCP Docker image is TBD
    - Built in brew by OSBS
    - brew package name: `ose-metering-helm-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70736
- presto
  - Written in Java, uses [maven][maven] as a project build tool.
  - origin
    - Source is https://github.com/operator-framework/presto/tree/master
    - Dockerfile is https://github.com/operator-framework/presto/blob/master/Dockerfile
    - Docker image is [quay.io/openshift/origin-metering-presto](https://quay.io/repository/openshift/origin-metering-presto)
    - Build by origin CI using prow/ci-operator
  - OCP
    - Source is http://pkgs.devel.redhat.com/cgit/containers/presto/
    - OCP Dockerfile is https://github.com/operator-framework/presto/blob/master/Dockerfile.rhel
    - OCP Docker image is TBD
    - Built in brew by OSBS
    - brew package name: `presto-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70772
- hive
  - Written in Java, uses [maven][maven] as a project build tool.
  - origin
    - Source is https://github.com/operator-framework/hive/tree/master
    - Dockerfile is https://github.com/operator-framework/hive/tree/master/Dockerfile
    - Docker image is [quay.io/openshift/origin-metering-hive](https://quay.io/repository/openshift/origin-metering-hive)
    - Build by origin CI using prow/ci-operator
  - OCP
    - Source is http://pkgs.devel.redhat.com/cgit/containers/hive/
    - OCP Dockerfile is https://github.com/operator-framework/hive/tree/master/Dockerfile.rhel
    - OCP Docker image is TBD
    - Built in brew by OSBS
    - brew package name: `hive-container`
    - url: https://brewweb.engineering.redhat.com/brew/packageinfo?packageID=70894
- hadoop
  - Written in Java, uses [maven][maven] as a project build tool.
  - origin
    - Source is https://github.com/operator-framework/hadoop/tree/master
    - Dockerfile is https://github.com/operator-framework/hadoop/tree/master/Dockerfile
    - Docker image is [quay.io/openshift/origin-metering-hadoop](https://quay.io/repository/openshift/origin-metering-hadoop)
    - Build by origin CI using prow/ci-operator
  - OCP
    - Source is http://pkgs.devel.redhat.com/cgit/containers/hadoop/
    - OCP Dockerfile is https://github.com/operator-framework/hadoop/tree/master/Dockerfile.rhel
    - OCP Docker image is TBD
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
  - Contains modified Dockerfiles: They use a tool [OIT][oit] that adds some additional changes to the repo before syncing it to dist-git.
  - One dist-git repo per docker image. Each repo is for one Dockerfile + other files used by OSBS.

For reporting-operator, helm, and metering-helm-operator, they all are written primarily in Go, and have any dependencies vendored, so OSBS can build them directly without any additional requirements once ART runs [oit][oit] to sync the repositories and update the Dockerfiles.

For Presto, Hive, and Hadoop, these are all written in Java, and their dependencies are fetched using [maven][maven], meaning extra work is required to build these, since OSBS doesn't allow downloading things outside the network.
To handle this we use [PNC][pnc] to perform builds using maven and push the artifacts to brew.

## Project New Castle (PNC)

Project New Castle (PNC) is basically an isolated build environment that proxies maven downloads ensuring downloads come from an internal maven repo.
It also proxies from maven central anything that isn't available yet, and uses persistent caching to ensure that once something is downloaded it doesn't change.
Once it's done the builds, we do a push to brew.

The Openshift Metering "product" is found at: http://orch.psi.redhat.com/pnc-web/#/product/50

Current product versions:

- Openshift Metering 0.7
  - Brew Tag Prefix: `openshift-metering-0.7-pnc`

We currently have 4 projects in PNC:

- presto
  - http://orch.psi.redhat.com/pnc-web/#/projects/207
  - build configs:
    - presto-0.212: http://orch.psi.redhat.com/pnc-web/#/projects/207/build-configs/571
    - presto-310: http://orch.psi.redhat.com/pnc-web/#/projects/207/build-configs/1470
- hive
  - http://orch.psi.redhat.com/pnc-web/#/projects/220
  - build configs:
    - hive-2.3.3: http://orch.psi.redhat.com/pnc-web/#/projects/220/build-configs/636
- hadoop
  - http://orch.psi.redhat.com/pnc-web/#/projects/219
  - build configs:
    - hadoop-3.1.1: http://orch.psi.redhat.com/pnc-web/#/projects/219/build-configs/633
- prometheus-jmx-exporter
  - http://orch.psi.redhat.com/pnc-web/#/projects/208
  - build configs:
    - prometheus-jmx-exporter-0.3.1: http://orch.psi.redhat.com/pnc-web/#/projects/208/build-configs/578

### PNC Brew Push

Read the PNC documentation on closing milestones: https://docs.engineering.redhat.com/pages/viewpage.action?pageId=44534467

Also read https://docs.engineering.redhat.com/display/JP/Integration+with+Brew for details on the brew tags that need to be created prior to the push.
Usually having tags added is as simple as filing a Jira ticket in the RCM project (Ex: https://projects.engineering.redhat.com/browse/RCM-43044).

[osbs]: https://osbs.readthedocs.io/en/latest/
[pnc]: https://docs.engineering.redhat.com/display/JP/User%27s+guide
[brew]: https://brewweb.engineering.redhat.com/brew/
[dist-git]: http://pkgs.devel.redhat.com/cgit/
[oit]: https://github.com/openshift/enterprise-images/blob/d44b833f8696102364f2526eaf130a961eb4cf56/oit.py
[maven]: https://maven.apache.org/
[pnc-brew]: https://docs.engineering.redhat.com/display/JP/Integration+with+Brew
