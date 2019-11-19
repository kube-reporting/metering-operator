# Testing OLM install with OCP content

This document is a summarization of https://docs.google.com/document/d/1t81RSsZbUoGO4r5OgJ1bqAESKt2fM25MvV6pcgQUPSk/edit#, please review this before proceeding as it covers how to request access to the necessary Quay organizations for pulling the OLM app bundles within Quay.io.

# Disable built-in OperatorSources in OCP 4.1

Disable ClusterVersionOperator management of the Marketplace redhat-operators OperatorSource so we can delete the existing one and install ours.

Store the following yaml in a file named `cvo-overrides.yaml`

```
apiVersion: config.openshift.io/v1
kind: ClusterVersion
metadata:
  name: version
spec:
  overrides:
  - kind: OperatorSource
    name: redhat-operators
    namespace: openshift-marketplace
    unmanaged: true
  - kind: OperatorSource
    name: community-operators
    namespace: openshift-marketplace
    unmanaged: true
```

Then run:

```
oc apply -f cvo-overrides.yaml
```

Delete the redhat-operators OperatorSource:

```
oc -n openshift-marketplace delete operatorsource redhat-operators
```

# Disable built-in OperatorSources in OCP 4.2

Store the following in a file called `operatorhub.yaml`:

```
apiVersion: config.openshift.io/v1
kind: OperatorHub
metadata:
  name: cluster
spec:
  disableAllDefaultSources: true
```

Then apply it:

```
oc apply -f operatorhub.yaml
```

# Get access to Quay organizations containing staged operator bundles

Add yourself to https://docs.google.com/spreadsheets/d/1OyUtbu9aiAi3rfkappz5gcq5FjUbMQtJG4jZCNqVT20/edit#gid=0 and get someone to grant you access.
This must be done before proceeding.
Once done, look at https://quay.io/application/ and verify you see the metering-ocp package listed in the registry namespaces `rh-operators-art` and `rh-osbs-operators`.

# Configure credentials

Next create a secret containing credentials containing your Quay credentials for accessing the app bundles ART builds.

Replace `$QUAY_AUTH_TOKEN` with the actual literal value of your `$QUAY_AUTH_TOKEN` and store this in a file named `marketplace-secret.yaml`

```
apiVersion: v1
kind: Secret
metadata:
  name: marketplacesecret
  namespace: openshift-marketplace
type: Opaque
stringData:
    token: "$QUAY_AUTH_TOKEN"
```

Next, create it:

```
oc apply -n openshift-marketplace -f marketplace-secret.yaml
```

# Configure operator source

Copy the following and store it in a file named `art-applications-operator-source.yaml`:

```
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: art-applications
  namespace: openshift-marketplace
spec:
  type: appregistry
  endpoint: https://quay.io/cnr
  # use "redhat-operators" to test production, current release contents.
  # use redhat-operators-art for pre-production content. This is unreleased, but for the current major release version.
  # use redhat-operators-stage for testing staging images.
  # use rh-verified-operators for the latest thing built (is the replacement of rh-osbs-operators) for stuff in the next major release.
  # rh-osbs-operators is the most regularly updated one.
  registryNamespace: redhat-operators-art
  # registryNamespace: redhat-operators-stage
  # registryNamespace: redhat-operators
  # registryNamespace: rh-verified-operators
  authorizationToken:
    secretName: marketplacesecret
```

Once you've copied and the file, install the operator source:

```
oc apply -n openshift-marketplace -f art-applications-operator-source.yaml
```

# Configure images to be mirrored

First, make sure you have [grpcurl](https://github.com/fullstorydev/grpcurl) installed, this will be used to query package information from the operator-registry pod containing our OLM package.

Next, open a port-forward to the `art-applications` operator-registry:

```
kubectl -n openshift-marketplace port-forward svc/art-applications 50051 &
```

Once we have the port-forward opened, we'll use the following script to print the images we're going to use, and then use eval on the output to set the environment variables we need so that our image mirroring script mirrors the correct content into the cluster:

```
hack/get-metering-package-images.sh
eval "$(hack/get-metering-package-images.sh)"
```

# Mirror images from Brew into your cluster's in-cluster image registry

Follow the [Mirroring OCP images into your cluster](mirroring-ocp-images.md) guide for instructions for mirroring images.
The previous step set some environment variables that the script will automatically use, so just follow the instructions to mirror your images into the cluster.

# Install

Proceed by following [Documentation/olm-install.md](../olm-install.md) and use `metering-ocp` instead of the other metering packages when searching for the package in the UI.
