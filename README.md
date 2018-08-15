# Admission Webhook
This project is used to explore the implementation of a Kubernetes [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks).

## Getting Started
To build the Docker image of the admission webhook server:
```
$ make image
```

To generate self-signed TLS artifacts:
```
# generate the self-signed CA
$ make tls/ca

# generate the TLS cert for the webhook server, signed by the custom CA
$ make tls/server
```

## References

1. [Diving Into Kubernetes MutatingAdmissionWebhook](https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74)
1. [Kubernetes External Admission Webhook Test Image](https://github.com/kubernetes/kubernetes/tree/v1.10.0-beta.1/test/images/webhook)
