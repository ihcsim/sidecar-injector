package webhook

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewClientset returns a new instance of Clientset. It authenticates with the cluster using the service account token mounted inside the webhook pod.
func NewClientset() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
