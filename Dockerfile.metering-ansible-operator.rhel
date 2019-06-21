# need the helm-cli from the helm image
FROM openshift/ose-metering-helm:latest as helm
# final image needs kubectl, so we copy `oc` from cli image to use as kubectl.
FROM openshift/ose-cli:latest as cli
# real base
FROM openshift/ose-ansible-operator:latest

USER root
RUN INSTALL_PKGS="curl bash ca-certificates less which inotify-tools tini python-boto3" \
    && yum install --setopt=skip_missing_names_on_install=False -y \
        $INSTALL_PKGS  \
    && yum clean all \
    && rm -rf /var/cache/yum

COPY --from=helm /usr/local/bin/helm /usr/local/bin/helm
COPY --from=cli /usr/bin/oc /usr/bin/oc
RUN ln -f -s /usr/bin/oc /usr/bin/kubectl

RUN yum -y update python-openshift

USER 1001
ENV HOME /opt/ansible
ENV HELM_CHART_PATH ${HOME}/charts/openshift-metering

COPY images/metering-ansible-operator/roles/ ${HOME}/roles/
COPY images/metering-ansible-operator/watches.yaml ${HOME}/watches.yaml
COPY images/metering-ansible-operator/scripts ${HOME}/scripts
COPY images/metering-ansible-operator/ansible.cfg /etc/ansible/ansible.cfg
COPY charts/openshift-metering ${HELM_CHART_PATH}

COPY manifests/deploy/openshift/olm/bundle /manifests

ENTRYPOINT ["/tini", "--", "/usr/local/bin/entrypoint"]

LABEL io.k8s.display-name="OpenShift metering-ansible-operator" \
      io.k8s.description="This is a component of OpenShift Container Platform and manages installation and configuration of all other metering components." \
      io.openshift.tags="openshift" \
      com.redhat.delivery.appregistry=true \
      maintainer="Chance Zibolski <sd-operator-metering@redhat.com>"
