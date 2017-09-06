package com.coreos.chargeback;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSCredentialsProvider;
import com.amazonaws.auth.BasicAWSCredentials;
import com.amazonaws.auth.BasicSessionCredentials;
import io.kubernetes.client.ApiException;
import java.io.IOException;
import java.net.URI;
import java.util.Map;
import org.apache.hadoop.conf.Configurable;
import org.apache.hadoop.conf.Configuration;

public class BucketSpecificCredentialsProvider implements AWSCredentialsProvider, Configurable {
  private static final String BUCKET_SECRET_NAMESPACE = "tectonic-chargeback";

  private String bucket;
  private Configuration conf;
  private BucketSecretClient client;
  private BucketSecretClient.BucketSecret secret;

  public BucketSpecificCredentialsProvider(URI uri, Configuration configuration) {
    bucket = uri.getHost();
    conf = configuration;
    try {
      client = new BucketSecretClient(BUCKET_SECRET_NAMESPACE);
    } catch (IOException e) {
      throw new RuntimeException("failed to configure Kubernetes client", e);
    }

    // TODO(DG): verify that this is the correct behavior
    refresh();
  }

  public AWSCredentials getCredentials() {
    if (secret == null) {
      throw new RuntimeException("credentials haven't been refreshed yet.");
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

  public void refresh() {
    Map<String, String> config;
    try {
      config = client.readBucketConfig();
    } catch (ApiException e) {
      throw new RuntimeException(
          String.format(
              "failed to read bucket config in ConfigMap '%s'", client.getConfigMapName()),
          e);
    }

    String secretName = config.get(bucket);
    if (secretName == null) {
      throw new IndexOutOfBoundsException(
          String.format(
              "configuration for bucket '%s' could not be found in ConfigMap '%s'",
              bucket, client.getConfigMapName()));
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
}
