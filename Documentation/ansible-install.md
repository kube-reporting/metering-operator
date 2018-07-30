# Installation using openshift-ansible

Using ansible is the recommend installation method for installing Metering on to an Openshift cluster.

The [openshift-metering playbook][metering-playbook] is located within the [openshift-ansible repo][openshift-ansible].

At the time this document was written, the openshift-metering playbook is available in the master branch and should be in the release-3.11 branch when it's created as well.

## Configuration

All of the supported configuration options are documented in [configuring metering][configuring-metering].
To supply custom configuration options set the `openshift_metering_config` variable to a dictionary containing the contents of the `Metering` `spec` field you wish to set.

For example:

```
openshift_metering_config:
  metering-operator:
    config:
      awsAccessKeyID: "REPLACEME"
```

## Install

Installing using the openshift-metering playbook will install the metering operator and it's components into the `openshift-metering` namespace.

First, clone openshift-ansible:

```
git clone https://github.com/openshift/openshift-ansible
```

Next, set the `openshift_metering_install` variable to true:

```yaml
openshift_metering_install: true
```

Finally, run the playbook:

```
ansible-playbook playbooks/openshift-metering/config.yml
```

If you're on GCP use the following instead:

```bash
ansible-playbook playbooks/openshift-metering/gcp-config.yml
```

### Verifying Operation and Metering Usage

Once you've installed Metering, make sure you set your `METERING_NAMESPACE` environment variable to `openshift-metering` and then return to the [verifying operation section][verifying-operation] of the main install doc.

After you've verified operation, continue on to [using Operator Metering][using-metering].

## Uninstall

Uninstall just requires setting `openshift_metering_install` to false, and re-running the `ansible-playbook` command from above:

```yaml
openshift_metering_install: false
```

[configuring-metering]: metering-config.md
[openshift-ansible]: https://github.com/openshift/openshift-ansible
[metering-playbook]: https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-metering
[verifying-operation]: install-metering.md#verifying-operation
[using-metering]: using-metering.md
