# Sidecar Injector

[![Codefresh build status]( https://g.codefresh.io/api/badges/pipeline/ihcsim/ihcsim%2Fsidecar-injector%2Fsidecar-injector?branch=master&type=cf-1)]( https://g.codefresh.io/repositories/ihcsim/sidecar-injector/builds?filter=trigger:build;branch:master;service:5b7e1d9e5904b8378771a864~sidecar-injector)

This project implements a Kubernetes [admission webhook](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#admission-webhooks) that injects a Nginx sidecar container to all pods on-creation.

* [Getting Started](#getting-started)
* [TLS](#tls)
* [References](#references)

This project differs from the [sample](https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74) in the following ways:

* A self-signed CA root certificate is added to the `MutatingAdmissionConfiguration` resource.
* All TLS artifacts are created and signed without relying on the k8s APIs. See the [Makefile](Makefile).
* The self-signed TLS cert defines the subject alternate names of the service DNS name in the [tls/san.cnf](tls/san.cnf) file.
* Use a new service account with only the necessary role.
* Unit tests are included.

## Getting Started
To generate self-signed TLS artifacts, run:
```
# generate the self-signed CA
$ make tls/ca

# generate the TLS cert for the webhook server, signed by the custom CA
$ make tls/server
```
This will create the self-signed CA cert and private key, and the server CSR, private key and cert in the `tls` folder.

To build the webhook server, push it to an image registry and deploy to Kubernetes, run:
```
$ IMAGE_REPO=<your_image_repo> make build push deploy
```
To enabled logging, specify the `DEBUG_ENABLED=true` environment variable.

For `deploy` to work, your local kubeconfig must be present.

Further testing with a busybox pod:
```
$ kubectl run busybox --image busybox --restart Never --command -- sleep 3600
pod "busybox" created

$ kubectl get po
NAME                               READY     STATUS    RESTARTS   AGE
busybox                            2/2       Running   0          1m
sidecar-injector-bd7d67dfb-5rpd8   1/1       Running   0          3m

$ kubectl describe po busybox
...
Containers:
  busybox:
    Container ID:  docker://a754cde1efaf6eefd081cc5cf9d625b35008d8b5692a2bce1a50b039ed13cdad
    Image:         busybox
    Image ID:      docker-pullable://busybox@sha256:cb63aa0641a885f54de20f61d152187419e8f6b159ed11a251a09d115fdff9bd
    Port:          <none>
    Host Port:     <none>
    Command:
      sleep
      3600
    State:          Running
      Started:      Wed, 22 Aug 2018 19:15:26 -0700
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-8p52s (ro)
  nginx:
    Container ID:   docker://9d9e558be64ecbcf6d05cb0a9d529c7d4d11be43346553170f61f01f2f0a9b8d
    Image:          nginx
    Image ID:       docker-pullable://nginx@sha256:d85914d547a6c92faa39ce7058bd7529baacab7e0cd4255442b04577c4d1f424
    Port:           80/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Wed, 22 Aug 2018 19:15:41 -0700
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:         <none>
...
```



To execute the unit tests, run:
```
$ make test
```

## TLS
All the TLS artifacts in the `tls` folder are self-signed samples.

The CA cert and private key used to sign the webhook server's cert are found in the `tls/ca` folder. The CA cert is also added to the `caBundle` field of the `MutatingWebhookConfiguration` resource. This will be used by the API Server to validate the webhook server's TLS cert.

The webhook server cert and private key are located in the `tls/server` folder. They are injected into the server container as `secret` resources. Note that the `service` names of the webhook server are added to the server cert as subject alternate names, specify in the `tls/san.cnf` file.

## References

1. [Diving Into Kubernetes MutatingAdmissionWebhook](https://medium.com/ibm-cloud/diving-into-kubernetes-mutatingadmissionwebhook-6ef3c5695f74)
1. [Kubernetes External Admission Webhook Test Image](https://github.com/kubernetes/kubernetes/tree/v1.10.0-beta.1/test/images/webhook)
