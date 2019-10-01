# Testing Metering OLM install with local manifests

This document covers how to test an OLM based install using local OLM manifests and your own Quay.io app repository.

The main reason you may want to do do this is if your making changes to the OLM manifests, and want to verify the behavior of the changes.

# Pre-requisites

You must have a [Quay.io](https://quay.io) account, if you do not, sign up for one.

# Setup

Make sure you have [operator-courier 2.1.7 or newer](https://github.com/operator-framework/operator-courier).

```
$ pip3 install --upgrade operator-courier
Collecting operator-courier
  Downloading https://***REMOVED***les.pythonhosted.org/packages/8a/3b/c8f3d95ee79a2d4992895b715095fcadeca7145f0b8fd7e5b9dd0ceecf24/operator_courier-2.1.7-py3-none-any.whl
Requirement already satis***REMOVED***ed, skipping upgrade: semver in /usr/local/lib/python3.7/site-packages (from operator-courier) (2.8.1)
Requirement already satis***REMOVED***ed, skipping upgrade: validators in /usr/local/lib/python3.7/site-packages (from operator-courier) (0.12.4)
Requirement already satis***REMOVED***ed, skipping upgrade: pyyaml in /usr/local/lib/python3.7/site-packages (from operator-courier) (5.1)
Requirement already satis***REMOVED***ed, skipping upgrade: requests in /usr/local/lib/python3.7/site-packages (from operator-courier) (2.21.0)
Requirement already satis***REMOVED***ed, skipping upgrade: six>=1.4.0 in /usr/local/Cellar/protobuf/3.7.1/libexec/lib/python3.7/site-packages (from validators->operator-courier) (1.12.0)
Requirement already satis***REMOVED***ed, skipping upgrade: decorator>=3.4.0 in /usr/local/lib/python3.7/site-packages (from validators->operator-courier) (4.4.0)
Requirement already satis***REMOVED***ed, skipping upgrade: chardet<3.1.0,>=3.0.2 in /usr/local/lib/python3.7/site-packages (from requests->operator-courier) (3.0.4)
Requirement already satis***REMOVED***ed, skipping upgrade: urllib3<1.25,>=1.21.1 in /usr/local/lib/python3.7/site-packages (from requests->operator-courier) (1.24.1)
Requirement already satis***REMOVED***ed, skipping upgrade: idna<2.9,>=2.5 in /usr/local/lib/python3.7/site-packages (from requests->operator-courier) (2.8)
Requirement already satis***REMOVED***ed, skipping upgrade: certi***REMOVED***>=2017.4.17 in /usr/local/lib/python3.7/site-packages (from requests->operator-courier) (2019.3.9)
Installing collected packages: operator-courier
  Found existing installation: operator-courier 2.1.4
    Uninstalling operator-courier-2.1.4:
      Successfully uninstalled operator-courier-2.1.4
Successfully installed operator-courier-2.1.7
```

Next, setup authentication for operator-courier by following the instructions at https://github.com/operator-framework/operator-courier#authentication.

# Pushing the bundle

Run the following commands, replacing `$QUAY_USERNAME` with your Quay.io username.

```
export QUAY_AUTH_TOKEN=$AUTH_TOKEN
./hack/push-olm-manifests.sh $QUAY_USERNAME metering-ocp manifests/deploy/openshift/olm/bundle
```

# Con***REMOVED***gure operator source

Copy the following and store it in a ***REMOVED***le named `metering-operator-source.yaml`, and replace `$QUAY_USERNAME` with your Quay.io username.

```
apiVersion: operators.coreos.com/v1
kind: OperatorSource
metadata:
  name: metering-operators-ocp-testing
  namespace: openshift-marketplace
spec:
  displayName: Metering Custom
  endpoint: https://quay.io/cnr
  publisher: Metering Developers
  registryNamespace: $QUAY_USERNAME
  type: appregistry
```

Once you've copied and modi***REMOVED***ed the ***REMOVED***le, install the operator source:

```
oc create -n openshift-marketplace -f metering-operator-source.yaml
```

# Install

You have two options for installation, you can use the Openshift OperatorHub UI or you can do it using the kubernetes CLI.

## Install Metering via Openshift OperatorHub UI

The process for testing is the same as what's documented in [Documentation/olm-install.md](../olm-install.md), the main difference is the package you will want to search for and use.

Follow the existing olm-install documentation, but before searching the OperatorHub UI, tick the *Custom* box underneath the search box to ***REMOVED***lter to packages provided via custom OperatorSources, like the one we just created.
The metering package you see should have an Operator Version of "4.3.0", with a Provider Type of "Custom".

## Install Metering using the CLI

The process for testing using the CLI is the same as what's documented in [Documentation/manual-olm-install.md](../manual-olm-install.md).
