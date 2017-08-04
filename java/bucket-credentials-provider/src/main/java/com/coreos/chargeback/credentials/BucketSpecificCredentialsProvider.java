package com.coreos.chargeback.credentials;

import com.amazonaws.auth.AWSCredentials;
import com.amazonaws.auth.AWSCredentialsProvider;
import org.apache.hadoop.conf.Configurable;
import org.apache.hadoop.conf.Configuration;
import java.net.URI;

public class BucketSpecificCredentialsProvider implements AWSCredentialsProvider, Configurable {
    public BucketSpecificCredentialsProvider(URI uri, Configuration configuration) {

    }

    public AWSCredentials getCredentials() {
        return null;
    }

    public void refresh() {

    }

    public void setConf(Configuration configuration) {

    }

    public Configuration getConf() {
        return null;
    }
}
