package com.coreos.chargeback;

import static org.junit.Assert.*;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSCredentialsProvider;
import java.net.URI;
import java.net.URISyntaxException;
import java.util.Map;
import org.apache.hadoop.conf.Configuration;
import org.junit.BeforeClass;
import org.junit.Test;

public class KubernetesCredentialsProviderTest {
  private static final String S3_ENV = "TEST_S3_BUCKET";

  private static String bucket;

  @BeforeClass
  public static void setup() {
    Map<String, String> env = System.getenv();
    bucket = env.get(S3_ENV);
  }

  @Test
  public void testAuthenticateActual() {
    if (bucket == null) {
      fail(String.format("A bucket must be set in the variable '%s'.", S3_ENV));
    }
    URI uri;
    try {
      uri = new URI(String.format("s3a://%s/billingStatement1", bucket));
    } catch (URISyntaxException e) {
      fail(
          String.format(
              "A validly formatted bucket name must be provided in %s: %s", S3_ENV, e.toString()));
      return;
    }

    AWSCredentialsProvider provider = getProvider(uri, new Configuration());
    AWSCredentials creds = provider.getCredentials();

    System.out.println("Found AWS Access Key: " + creds.getAWSAccessKeyId());
  }

  public static AWSCredentialsProvider getProvider(URI uri, Configuration conf) {
    return new KubernetesCredentialsProvider(uri, conf);
  }
}
