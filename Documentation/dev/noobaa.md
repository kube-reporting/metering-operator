# Installing NooBaa

Follow the instructions on https://github.com/noobaa/noobaa-operator/ for installing NooBaa.

If you have an existing noobaa install, the management UI and S3 connection information can be found by running `noobaa status`.

It is highly recommended that before beginning you familiarize yourself with NooBaa.
If you have not already, you should: con***REMOVED***gure a storage pool using the NooBaa management UI, con***REMOVED***gure your Tier 1 data placement, and create a bucket for usage by Metering.

## Con***REMOVED***guring Metering for NooBaa

Currently noobaa supports TLS, but produces a self-signed cert that we cannot trust due to it not being signed for any of the hostnames we connect to it using.
You have a few choices:

- Do not use TLS to connect to Noobaa (use the http:// protocol and port 80 instead of https:// with 443).
- Con***REMOVED***gure a CA for NooBaa and ensure that `s3.noobaa.svc` is in the subject alternate names of the server certi***REMOVED***cate.
  - There is an [open issue](https://github.com/noobaa/noobaa-operator/issues/43#issue-487679497) requesting that certi***REMOVED***cate generation is handled automatically on Openshift, and contains instructions on how you can create a certi***REMOVED***cate in Openshift.

### Create the namespace

Before we can start, we need to create the namespace so that we can create the secret containing our NooBaa credentials:

```
kubectl create ns $METERING_NAMESPACE
```

### Create noobaa credentials secret

Run the following command to print our your credentials:

```
noobaa status 2>&1 | grep AWS_
```

Create a secret storing your NooBaa AWS credentials:

```
kubectl -n $METERING_NAMESPACE create secret generic my-noobaa-secret --from-literal=aws-access-key-id=your-access-key  --from-literal=aws-secret-access-key=your-secret-key
```

### MeteringCon***REMOVED***g

Below are two example `MeteringCon***REMOVED***g` resources you can use when trying to use Metering with NooBaa.
It is recommended you create a bucket in the NooBaa UI or using the AWS CLI dedicated for Metering and change the value of `spec.storage.hive.s3Compatible.bucket` from `***REMOVED***rst.bucket` to the name of the bucket you created.

### Without TLS

When using NooBaa without TLS you need to set the `spec.storage.hive.s3Compatible.endpoint` value to `http://s3.noobaa.svc:80`.

```
apiVersion: metering.openshift.io/v1
kind: MeteringCon***REMOVED***g
metadata:
  name: "operator-metering"
spec:
  storage:
    type: "hive"
    hive:
      type: "s3Compatible"
      s3Compatible:
        bucket: "***REMOVED***rst.bucket"
        endpoint: "http://s3.noobaa.svc:80"
        secretName: "my-noobaa-secret"
```

### With TLS

When using NooBaa with TLS you need to set the `spec.storage.hive.s3Compatible.endpoint` value to `https://s3.noobaa.svc:443` and set the `spec.storage.hive.s3Compatible.ca` ***REMOVED***elds.

```
apiVersion: metering.openshift.io/v1
kind: MeteringCon***REMOVED***g
metadata:
  name: "operator-metering"
spec:
  storage:
    type: "hive"
    hive:
      type: "s3Compatible"
      s3Compatible:
        bucket: "***REMOVED***rst.bucket"
        endpoint: "https://s3.noobaa.svc:443"
        secretName: "my-noobaa-secret"
        ca:
          createSecret: true
          # You MUST replace the value below with the certi***REMOVED***cate chain for
          # your NooBaa installation.
          content: |
            -----BEGIN CERTIFICATE-----
            MIIDdTCCAl2gAwIBAgIIOT6yX8oQ/S4wDQYJKoZIhvcNAQELBQAwNjE0MDIGA1UE
            Awwrb3BlbnNoaWZ0LXNlcnZpY2Utc2VydmluZy1zaWduZXJAMTU2NDAwMzI4MTAe
            Fw0xOTA4MzAyMDQ5MzlaFw0yMTA4MjkyMDQ5NDBaMBgxFjAUBgNVBAMTDXMzLm5v
            b2JhYS5zdmMwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDhYlylOSHB
            ++XGfsuyFLvK8Yrri2d0NRJBBqmQ4c1Mhv242HSVMxyRsxBXyLnomjvFiKJhMlIB
            m/LJRxBJz3ZONvkCiBF8SOmim6TK+9I8Ky7+nz8urovJY28jiJjjsLAJitk/nU40
            boZ41BD4oTVVmY7u2iB7JyurTGZejSYGf1rq1cHeNOZ2PPHpsIFvfvcNh4ig6lOj
            Jfz9shaG7P4zDRK7vZVJylmLDq2s1oauhDs57QCLMBMmxrm96+NdRP7BiV1qmpwZ
            QavLTQcLNMOFGgN8ETJ/nT85zd8tHKePL0tL/KULVfo18rHjcYTUyLManZCXrGdG
            HGmfsOBbtkJLAgMBAAGjgaQwgaEwDgYDVR0PAQH/BAQDAgWgMBMGA1UdJQQMMAoG
            CCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwNQYDVR0RBC4wLIINczMubm9vYmFhLnN2
            Y4IbczMubm9vYmFhLnN2Yy5jbHVzdGVyLmxvY2FsMDUGCysGAQQBkggRZAIBBCYT
            JDE3YzdiZWYzLWJkM2YtMTFlOS04ODUxLTAyMDVlYmEwYTZiNjANBgkqhkiG9w0B
            AQsFAAOCAQEAajW7mCp7S//NxJGaJUrH+08zV5Q8PzdFWqnZ6k3ZpyvqLmIiV0VZ
            2YQtyd+SxyekIbgYXHHhrUPFKL/coUGzHqjw/F+ZvysShsIvzHyFKyMXP1Zc7WeU
            83PLjjReNHv7iII62/wCPdYIFr1dNFPnfQaSrIcrN+OyiH4FVQd187BArBkudSBw
            Y7gGq8XI80IAutbxnYGgtElKOrbh8MELlrPfqMlI+1/U0upP+AEde78LpDgTnI2H
            16gNjkM+CDOgWN2njqdiohI42Uo3uan4LVTB07FOp5ulB9TmlAn1HcpFc1YQEiVo
            uvbyiXoRx34oaSKNwXBA5QYwYh8A3a74dA==
            -----END CERTIFICATE-----
            -----BEGIN CERTIFICATE-----
            MIIDCjCCAfKgAwIBAgIBATANBgkqhkiG9w0BAQsFADA2MTQwMgYDVQQDDCtvcGVu
            c2hpZnQtc2VydmljZS1zZXJ2aW5nLXNpZ25lckAxNTY0MDAzMjgxMB4XDTE5MDcy
            NDIxMjEyMFoXDTIwMDcyMzIxMjEyMVowNjE0MDIGA1UEAwwrb3BlbnNoaWZ0LXNl
            cnZpY2Utc2VydmluZy1zaWduZXJAMTU2NDAwMzI4MTCCASIwDQYJKoZIhvcNAQEB
            BQADggEPADCCAQoCggEBAKzp3pWqhEGqOXWTu7GwegQC43L+IXwgy9ROF8aesLV5
            WqVVIG5asg4APBVOw8bQUIKUTAEMC6Z8uhY8/H5UVVgspLPoyfHkP0Fiza8uwFB7
            dy3f37F7LAgiopAFRvVLbfz0jo0s6HY3PuOQnzv0+K3Fqvp6/VrJXiUvhNI5Q2ic
            ifSJ+f/hg/KVEf2vezreT91e+ijQVbTEPsnV3y24DnDAVo3gnuXON89DtLszcXHG
            T/AXX3ca/hh9kFExwV6y3/ahbX3a7f0hlrLDQ4ifmqnIhibI3+SsjdSj98rk77yl
            YfKNomCr4AV1zvCMB7ssBStdyqIpjF1afflrhdKI9+cCAwEAAaMjMCEwDgYDVR0P
            AQH/BAQDAgKkMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBACo6
            Vwx1yqd698UvxHfTpwzmKl049zG1S4coMiRXBemzaU50U/Twj2n4fhN8ogh6FNsN
            4usl+4PimrYvACnteQ7ICsEkfOWCugpwH8N7o66NsrqTMe9micSuDXr0lRCimmcw
            pFIoTWC1sonOT+7vESQJccB4BE93hwQD1ZRriAWAVtdxG1rqN+vkEdW6DoRHfU5i
            7PrDmMgrXl3vV2gTQuLtAvWeSTZe/4bWGZkqhYBOvPNNaGQbmU3y7To6L8BKoa1+
            14EsTPkbKPDa4bTnFbS0x4FguHa8vIyF8uAkU6mnExzjcJCRZqBakfK/17qdqzcR
            MxX6xP1czY3Cop0HoaY=
            -----END CERTIFICATE-----
```

## Install

Proceed with the [installation documentation](../install-metering.md).
