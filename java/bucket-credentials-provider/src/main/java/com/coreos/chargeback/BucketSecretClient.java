package com.coreos.chargeback;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.V1ConfigMap;
import io.kubernetes.client.models.V1Secret;
import io.kubernetes.client.util.Config;
import java.io.IOException;
import java.util.Map;

public class BucketSecretClient extends CoreV1Api {
  // stores mapping from bucket name to secret name
  private static final String BUCKET_CONFIGMAP_NAME = "buckets";

  // Provider specific configuration keys
  private static final String AWS_ID_STR = "AWS_ACCESS_KEY_ID";
  private static final String AWS_KEY_STR = "AWS_SECRET_ACCESS_KEY";
  private static final String AWS_SESSION_STR = "AWS_SESSION_TOKEN";
  private static final String AWS_CRED_PROVIDER_STR = "CredentialsProvider";

  String namespace;

  public class BucketSecret {
    String AWSAccessKeyID;
    String AWSSecretAccessKey;
    String AWSSessionToken;
    String AWSCredentialsProvider;
  }

  public BucketSecretClient(String namespace) throws IOException {
    this(Config.defaultClient(), namespace);
  }

  public BucketSecretClient(ApiClient client, String namespace) {
    super(client);
    this.namespace = namespace;
  }

  public BucketSecret readBucketSecret(String secretName) throws ApiException {
    V1Secret secret = this.readNamespacedSecret(secretName, this.namespace, "", false, true);
    Map<String, String> data = secret.getStringData();

    BucketSecret bucketSecret = new BucketSecret();
    bucketSecret.AWSAccessKeyID = data.get(AWS_ID_STR);
    bucketSecret.AWSSecretAccessKey = data.get(AWS_KEY_STR);
    bucketSecret.AWSSessionToken = data.get(AWS_SESSION_STR);
    bucketSecret.AWSCredentialsProvider = data.get(AWS_CRED_PROVIDER_STR);
    return bucketSecret;
  }

  public Map<String, String> readBucketConfig() throws ApiException {
    V1ConfigMap config =
        this.readNamespacedConfigMap(BUCKET_CONFIGMAP_NAME, this.namespace, "", false, true);
    return config.getData();
  }

  public String getConfigMapName() {
    return String.format("%s/%s", this.namespace, BUCKET_CONFIGMAP_NAME);
  }
}
