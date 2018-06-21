{{- define "hive-env" }}
        - name: CORE_CONF_fs_defaultFS
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: default-fs
              optional: true
        - name: CORE_CONF_fs_s3a_access_key
          valueFrom:
            secretKeyRef:
              name: presto-secrets
              key: aws-access-key-id
        - name: CORE_CONF_fs_s3a_secret_key
          valueFrom:
            secretKeyRef:
              name: presto-secrets
              key: aws-secret-access-key
        - name: HIVE_SITE_CONF_hive_metastore_uris
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: metastore-uris
        - name: HIVE_SITE_CONF_javax_jdo_option_ConnectionURL
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: db-connection-url
        - name: HIVE_SITE_CONF_javax_jdo_option_ConnectionDriverName
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: db-connection-driver
        - name: HIVE_SITE_CONF_javax_jdo_option_ConnectionUserName
          valueFrom:
            secretKeyRef:
              name: hive-common-secrets
              key: db-connection-username
              optional: true
        - name: HIVE_SITE_CONF_javax_jdo_option_ConnectionPassword
          valueFrom:
            secretKeyRef:
              name: hive-common-secrets
              key: db-connection-password
              optional: true
        - name: HIVE_SITE_CONF_hive_metastore_schema_verification
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: enable-metastore-schema-verification
        - name: HIVE_SITE_CONF_datanucleus_schema_autoCreateAll
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: auto-create-metastore-schema
        - name: HIVE_SITE_CONF_hive_default_fileformat
          valueFrom:
            configMapKeyRef:
              name: hive-common-config
              key: default-file-format
        - name: MY_NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        - name: MY_POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: MY_POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: JAVA_MAX_MEM_RATIO
          value: "50"
{{- end }}
