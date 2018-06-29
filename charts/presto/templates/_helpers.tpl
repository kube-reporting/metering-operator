{{- de***REMOVED***ne "hive-env" }}
        - name: CORE_CONF_fs_defaultFS
          valueFrom:
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
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
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
              key: metastore-uris
        - name: HIVE_SITE_CONF_javax_jdo_option_ConnectionURL
          valueFrom:
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
              key: db-connection-url
        - name: HIVE_SITE_CONF_javax_jdo_option_ConnectionDriverName
          valueFrom:
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
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
        - name: HIVE_SITE_CONF_hive_metastore_schema_veri***REMOVED***cation
          valueFrom:
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
              key: enable-metastore-schema-veri***REMOVED***cation
        - name: HIVE_SITE_CONF_datanucleus_schema_autoCreateAll
          valueFrom:
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
              key: auto-create-metastore-schema
        - name: HIVE_SITE_CONF_hive_default_***REMOVED***leformat
          valueFrom:
            con***REMOVED***gMapKeyRef:
              name: hive-common-con***REMOVED***g
              key: default-***REMOVED***le-format
        - name: MY_NODE_NAME
          valueFrom:
            ***REMOVED***eldRef:
              ***REMOVED***eldPath: spec.nodeName
        - name: MY_POD_NAME
          valueFrom:
            ***REMOVED***eldRef:
              ***REMOVED***eldPath: metadata.name
        - name: MY_POD_NAMESPACE
          valueFrom:
            ***REMOVED***eldRef:
              ***REMOVED***eldPath: metadata.namespace
        - name: JAVA_MAX_MEM_RATIO
          value: "50"
{{- end }}
