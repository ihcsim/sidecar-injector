package webhook

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	patchPathContainer  = "/spec/containers/-"
	patchPathAnnotation = "/metadata/annotations"
)

// PodPatch represents a RFC 6902 patch document for pods.
type PodPatch struct {
	original *corev1.Pod
	patchOps []*patchOp
}

// NewPodPatch returns a new instance of PodPatch.
func NewPodPatch(pod *corev1.Pod) *PodPatch {
	return &PodPatch{
		original: pod,
		patchOps: []*patchOp{},
	}
}

func (p *PodPatch) addContainerPatch(container *corev1.Container) {
	p.patchOps = append(p.patchOps, &patchOp{
		Op:    "add",
		Path:  patchPathContainer,
		Value: container,
	})
}

func (p *PodPatch) addAnnotationPatch() {
	p.patchOps = append(p.patchOps, &patchOp{
		Op:    "add",
		Path:  patchPathAnnotation,
		Value: map[string]string{annotationKeySidecarInjection: "false"},
	})
}

// patchOp represents a RFC 6902 patch operation.
type patchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:value,omitempty`
}
