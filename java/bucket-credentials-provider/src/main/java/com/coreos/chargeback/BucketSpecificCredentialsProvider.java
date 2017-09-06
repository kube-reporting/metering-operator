package com.coreos.chargeback;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSCredentialsProvider;
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
    // use bucket name to retrieve secret name and class name from ConfigMap
    // if credentialsProvider is given then load the class and delegate to it
    // if secret is given, retrieve it's contents from an API server
    // If no credentialsProvider and no secret are given then use the default chaining provider
    // otherwise:
    //  - ensure secret has "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"
    //  - if has "AWS_SESSION_TOKEN" use static sts provider, otherwise use static credentials
    // provider
    return null;
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
