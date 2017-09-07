package com.coreos.chargeback;

import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.V1Secret;
import io.kubernetes.client.util.Config;
import java.io.IOException;
import java.util.Map;

/** Extends Kubernetes API with Chargeback specific configuration types. */
public class BucketSecretClient extends CoreV1Api {
  // Provider specific configuration keys
  private static final String AWS_ID_STR = "AWS_ACCESS_KEY_ID";
  private static final String AWS_KEY_STR = "AWS_SECRET_ACCESS_KEY";
  private static final String AWS_SESSION_STR = "AWS_SESSION_TOKEN";
  private static final String AWS_CRED_PROVIDER_STR = "CredentialsProvider";

  /** Namespace all resources should be retrieved from. */
  String namespace;

  /** Credentials configuration that is stored in Kubernetes as a Secret. */
  public class BucketSecret {
    String AWSAccessKeyID;
    String AWSSecretAccessKey;
    String AWSSessionToken;
    String AWSCredentialsProvider;
  }

  /**
   * Uses default Kubernetes client configuration discovery mechanisms.
   *
   * @param namespace The namespace used for all requests.
   * @throws IOException when client cannot be configured
   */
  public BucketSecretClient(String namespace) throws IOException {
    this(Config.defaultClient(), namespace);
  }

  /**
   * Configure API with given Kubernetes client.
   *
   * @param client Client used to make requests.
   * @param namespace The namespace used for all requests.
   */
  public BucketSecretClient(ApiClient client, String namespace) {
    super(client);
    this.namespace = namespace;
  }

  /**
   * Retrieves Chargeback credentials stored in a Kubernetes Secret.
   *
   * @param secretName Name of the secret to use
   * @return Bucket specific credentials configuration.
   * @throws ApiException if cannot talk to API server or Secret does not exist
   */
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
}
