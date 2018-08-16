# Admission Webhook
I used this project to explore the implementation of a Kubernetes [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks).

## Getting Started
To build the Docker image of the admission webhook server, run:
```
$ make build
```

To generate self-signed TLS artifacts, run:
```
# generate the self-signed CA
$ make tls/ca

# generate the TLS cert for the webhook server, signed by the custom CA
$ make tls/server
```

To deploy to Kubernetes, run:
```
$ make deploy
```

To enabled debug mode, run:
```
$ DEBUG_ENABLED=true make deploy
```

## TLS
All the TLS artifacts in the `tls` folder are self-signed examples.

The CA cert and private key used to sign the webhook server's TLS cert are in the `tls/ca` folder. The CA cert is also added to the `caBundle` field of the `MutatingWebhookConfiguration` resource. This will be used by the API Server to validate the webhook server's TLS cert.

The webhook server cert and private key are in the `tls/server` folder. They are injected into to the server container as `Secret` resources. Note that the `Service` names of the webhook server are added as subject alternate names to the server cert.

## References

1. [Diving Into Kubernetes MutatingAdmissionWebhook](https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74)
1. [Kubernetes External Admission Webhook Test Image](https://github.com/kubernetes/kubernetes/tree/v1.10.0-beta.1/test/images/webhook)
