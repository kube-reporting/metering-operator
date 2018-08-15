# Manual Installation

If you want to install metering without OLM, using what's currently in master, ***REMOVED***rst start by cloning the repo.
Next, decide which namespace you want to install Metering into, and set the `METERING_NAMESPACE` environment variable to the namespace you want to use.
By default, if it's unset, it will use the `metering` namespace.

## Install

Depending on your Kubernetes platform (regular Kubernetes, Tectonic, or Openshift)

For a standard Kubernetes cluster:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/install.sh
```

If your using Tectonic, use tectonic-install.sh:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/tectonic-install.sh
```

If your using Openshift, use openshift-install.sh:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/openshift-install.sh
```

## Uninstall

To uninstall the process is the same, pick the right uninstall script for your platform, and run it.

For a standard Kubernetes cluster:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/uninstall.sh
```

If your using Tectonic, use tectonic-uninstall.sh:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/tectonic-uninstall.sh
```

If your using Openshift, use openshift-uninstall.sh:

```
$ export METERING_NAMESPACE=metering-$USER
$ ./hack/openshift-uninstall.sh
```

## Customize installation


If you wish to customize the installation, such as to modify con***REMOVED***guration
options, change the image tag or repository, then you can use a custom
`metering` resource. To start, copy the default metering resource to a
separate ***REMOVED***le that we can modify:

```
$ cp manifests/metering-con***REMOVED***g/default.yaml metering-custom.yaml
```

For developers, the most common change is modifying the image tag, con***REMOVED***g, and resource limits.
Take a look at `manifests/metering-con***REMOVED***g/custom-values.yaml` to get an
idea of what you can modify that relates to con***REMOVED***guration and resource limits, and
`manifests/metering-con***REMOVED***g/latest-versions.yaml` to see  how to change the
image tag of each component.

```
$ export METERING_NAMESPACE=metering-$USER
$ export METERING_CR_FILE=metering-custom.yaml
```

Then run the installation script for your platform:

- `./hack/install.sh`
- `./hack/tectonic-install.sh`
- `./hack/openshift-install.sh`

For more details on con***REMOVED***guration options, most are documented in the [con***REMOVED***guring metering document][con***REMOVED***guring-metering].

## Run metering operator locally

It's also possible to run the operator locally.
To simplify this, we've got a few `Make***REMOVED***le` targets to handle the building and running of the operator.

First, we still need to run Presto, Hive, and HDFS in the cluster, and also set reporting-operator replicas to 0 so that our local operator can obtain the leader election lease when we start it.

To do this, update your `metering-custom.yaml` to set `spec.reporting-operator.replicas` to `0` like so:

```
spec:
  reporting-operator:
    replicas: 0
```

Next, run the install script for your platform (see above).

After running the install script, ***REMOVED***gure out where your Prometheus pod is running.
By default the `run-reporting-operator-local` Make***REMOVED***le target assumes that the pod is in the `tectonic-system` namespace and can be queried using the label selector `app=prometheus`.

If you're Prometheus is located somewhere, ***REMOVED***, you can override the defaults using the environment variables `METERING_PROMETHEUS_NAMESPACE` and `METERING_PROMTHEUS_LABEL_SELECTOR` to the namespace your Prometheus pod is in, and the label selector for querying Prometheus.

Ex (these are the defaults):
```
export METERING_PROMETHEUS_NAMESPACE=tectonic-system
export METERING_PROMTHEUS_LABEL_SELECTOR=app=prometheus
```

Finally, use the following command to build & run the operator:

```
make run-reporting-operator-local
```

The above command builds the operator for your local OS (by default it only builds for Linux), uses kubectl port-forward to make Prometheus, Presto, and Hive available locally for your operator to communicate with, and then starts the operator with con***REMOVED***guration set to use these local port-forwards.
Lastly, the operator automatically uses your `$KUBECONFIG` to connect and authenticate to your cluster and perform Kubernetes API calls.


[con***REMOVED***guring-metering]: metering-con***REMOVED***g.md
