package test

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// FakeClient is a fake clientset that implements the kubernetes.Interface.
type FakeClient struct {
	kubernetes.Interface
}

// NewFakeClient returns a fake Kubernetes clientset.
func NewFakeClient() (kubernetes.Interface, error) {
	client := fake.NewSimpleClientset()
	return &FakeClient{client}, nil
}
