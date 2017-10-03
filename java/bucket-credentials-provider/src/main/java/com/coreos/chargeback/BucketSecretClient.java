package com.coreos.chargeback;

import static io.kubernetes.client.util.Config.*;

import com.google.gson.reflect.TypeToken;
import com.squareup.okhttp.Call;
import io.kubernetes.client.ApiClient;
import io.kubernetes.client.ApiException;
import io.kubernetes.client.ApiResponse;
import io.kubernetes.client.apis.CoreV1Api;
import io.kubernetes.client.models.V1ConfigMap;
import io.kubernetes.client.util.Config;
import java.io.File;
import java.io.FileInputStream;
import java.io.IOException;
import java.util.Base64;
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
    this(client(), namespace);
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
    Call call =
        this.readNamespacedSecretCall(secretName, this.namespace, "", false, true, null, null);
    // hack to workaround broken secret support
    ApiResponse<V1ConfigMap> secret =
        this.getApiClient().execute(call, new TypeToken<V1ConfigMap>() {}.getType());
    Map<String, String> data = secret.getData().getData();

    BucketSecret bucketSecret = new BucketSecret();
    bucketSecret.AWSAccessKeyID = decodeB64(data.get(AWS_ID_STR));
    bucketSecret.AWSSecretAccessKey = decodeB64(data.get(AWS_KEY_STR));
    bucketSecret.AWSSessionToken = decodeB64(data.get(AWS_SESSION_STR));
    bucketSecret.AWSCredentialsProvider = decodeB64(data.get(AWS_CRED_PROVIDER_STR));
    return bucketSecret;
  }

  /**
   * Attempt to use default configuration but fallback to vendored Pod service account logic.
   *
   * @return
   * @throws IOException
   */
  public static ApiClient client() throws IOException {
    try {
      return Config.defaultClient();
    } catch (RuntimeException e) {
      return inClusterClient();
    }
  }

  public static ApiClient inClusterClient() throws IOException {
    // API server URL
    String apiHost = System.getenv(ENV_SERVICE_HOST);
    String apiPort = System.getenv(ENV_SERVICE_PORT);
    String apiServer = String.format("https://%s:%s", apiHost, apiPort);

    // load token from service account
    String token = readFile(SERVICEACCOUNT_TOKEN_PATH);

    // read CA from disk
    FileInputStream caIn = new FileInputStream(SERVICEACCOUNT_CA_PATH);

    ApiClient client = fromUrl(apiServer);
    client.setApiKey("Bearer " + token);
    client.setSslCaCert(caIn);
    return client;
  }

  private static String readFile(String path) throws IOException {
    File f = new File(path);
    FileInputStream in = new FileInputStream(f);
    byte[] data = new byte[(int) f.length()];
    in.read(data);
    in.close();

    return new String(data, "UTF-8");
  }

  private static String decodeB64(String in) {
    if (in == null) {
      return null;
    }

    return new String(Base64.getDecoder().decode(in));
  }
}
