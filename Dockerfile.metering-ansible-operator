# need the helm-cli from the helm image
FROM quay.io/openshift/origin-metering-helm:latest as helm
# final image needs kubectl, so we copy `oc` from cli image to use as kubectl.
FROM openshift/origin-cli:latest as cli
# real base
FROM quay.io/operator-framework/ansible-operator:v0.6.0

USER root
RUN INSTALL_PKGS="curl bash ca-certificates less which inotify-tools" \
    && yum -y install epel-release \
    && yum install --setopt=skip_missing_names_on_install=False -y \
        $INSTALL_PKGS  \
    && yum clean all \
    && rm -rf /var/cache/yum

ENV TINI_VERSION v0.18.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini

COPY --from=helm /usr/local/bin/helm /usr/local/bin/helm
COPY --from=cli /usr/bin/oc /usr/bin/oc
RUN ln -f -s /usr/bin/oc /usr/bin/kubectl

RUN pip install --no-cache-dir --upgrade openshift
RUN pip install boto3

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
