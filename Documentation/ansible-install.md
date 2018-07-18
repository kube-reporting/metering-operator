# Installation using openshift-ansible

Using ansible is the recommend installation method for installing Metering on to an Openshift cluster.

The [openshift-metering playbook][metering-playbook] is located within the [openshift-ansible repo][openshift-ansible].

At the time this document was written, the openshift-metering playbook is available in the master branch and should be in the release-3.11 branch when it's created as well.

## Con***REMOVED***guration

All of the supported con***REMOVED***guration options are documented in [con***REMOVED***guring metering][con***REMOVED***guring-metering].
To supply custom con***REMOVED***guration options set the `openshift_metering_con***REMOVED***g` variable to a dictionary containing the contents of the `Metering` `spec` ***REMOVED***eld you wish to set.

For example:

```
openshift_metering_con***REMOVED***g:
  metering-operator:
    con***REMOVED***g:
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
ansible-playbook playbooks/openshift-metering/con***REMOVED***g.yml
```

If you're on GCP use the following instead:

```bash
ansible-playbook playbooks/openshift-metering/gcp-con***REMOVED***g.yml
```

## Uninstall

Uninstall just requires setting `openshift_metering_install` to false, and re-running the `ansible-playbook` command from above:

```yaml
openshift_metering_install: false
```

[con***REMOVED***guring-metering]: metering-con***REMOVED***g.md
[openshift-ansible]: https://github.com/openshift/openshift-ansible
[metering-playbook]: https://github.com/openshift/openshift-ansible/tree/master/playbooks/openshift-metering
