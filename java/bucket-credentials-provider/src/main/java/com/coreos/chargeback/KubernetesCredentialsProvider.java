package com.coreos.chargeback;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSCredentialsProvider;
import com.amazonaws.auth.BasicAWSCredentials;
import com.amazonaws.auth.BasicSessionCredentials;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.models.V1ConfigMap;
import java.io.IOException;
import java.net.URI;
import java.util.Map;
import org.apache.hadoop.conf.Configurable;
import org.apache.hadoop.conf.Configuration;

/**
 * KubernetesCredentialsProvider uses Kubernetes objects for storage of credentials and their
 * configuration. A ConfigMap stores a mapping of S3 buckets to Secret Kubernetes objects.
 * Credentials are minted by retrieving the Kubernetes Secret from the given bucket and using it's
 * contents to create the correct credentials.
 */
public class KubernetesCredentialsProvider implements AWSCredentialsProvider, Configurable {
  // namespace used for each Secret and ConfigMap
  private static final String BUCKET_SECRET_NAMESPACE = "tectonic-chargeback";

  // stores mapping from bucket name to secret name
  private static final String BUCKET_CONFIGMAP_NAME = "buckets";

  private String bucket;
  private Configuration conf;
  private BucketSecretClient client;
  private BucketSecretClient.BucketSecret secret;

  /**
   * Creates a credentials provider which uses the bucket requested to determine credentials used.
   *
   * @param uri identifier containing the S3 bucket name
   * @param configuration Hadoop config
   */
  public KubernetesCredentialsProvider(URI uri, Configuration configuration) {
    bucket = uri.getHost();
    conf = configuration;
    try {
      client = new BucketSecretClient(BUCKET_SECRET_NAMESPACE);
    } catch (IOException e) {
      throw new RuntimeException("failed to configure Kubernetes client", e);
    }
  }

  /**
   * Creates credentials based on stored Kubernetes state. If credentials have not been retrieved
   * yet, they will on the first invocation.
   *
   * @return Access key or STS AWS credentials
   */
  public AWSCredentials getCredentials() {
    if (secret == null) {
      refresh();
    }

    if (secret.AWSCredentialsProvider != null) {
      // TODO(DG): chain to other providers
      throw new UnsupportedOperationException(
          "specifying alternative providers not yet implemented.");
    }

    // use AWS STS if possible
    if (secret.AWSSessionToken != null) {
      return new BasicSessionCredentials(
          secret.AWSAccessKeyID, secret.AWSSecretAccessKey, secret.AWSSessionToken);
    }
    return new BasicAWSCredentials(secret.AWSAccessKeyID, secret.AWSSecretAccessKey);
  }

  /**
   * Using the Kubernetes API determines the Secret configured for the given bucket and retrieves
   * it.
   */
  public void refresh() {
    Map<String, String> config;
    try {
      config = readBucketConfig();
    } catch (ApiException e) {
      throw new RuntimeException(
          String.format("failed to read bucket config in ConfigMap '%s'", getConfigMapName()), e);
    }

    String secretName = config.get(bucket);
    if (secretName == null) {
      throw new IndexOutOfBoundsException(
          String.format(
              "configuration for bucket '%s' could not be found in ConfigMap '%s'",
              bucket, getConfigMapName()));
    }

    try {
      secret = client.readBucketSecret(secretName);
    } catch (ApiException e) {
      throw new RuntimeException(
          String.format("couldn't get bucket secret '%s/%s'", client.namespace, secretName), e);
    }
  }

  public void setConf(Configuration configuration) {
    conf = configuration;
  }

  public Configuration getConf() {
    return conf;
  }

  /**
   * Retrieve ConfigMap holding the S3 bucket credential configuration.
   *
   * @return mapping from S3 bucket name to Kubernetes Secret name
   * @throws ApiException if cannot talk to API server or ConfigMap does not exist
   */
  public Map<String, String> readBucketConfig() throws ApiException {
    V1ConfigMap config =
        client.readNamespacedConfigMap(BUCKET_CONFIGMAP_NAME, client.namespace, "", false, true);
    return config.getData();
  }

  /**
   * Identifies the ConfigMap object used by this instance to store credential configuration.
   *
   * @return string formatted '<namespace>'/'<name>'
   */
  public String getConfigMapName() {
    return String.format("%s/%s", client.namespace, BUCKET_CONFIGMAP_NAME);
  }
}
