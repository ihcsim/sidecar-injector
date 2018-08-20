package webhook

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ihcsim/admission-webhook/test"
)

func TestPodPatch(t *testing.T) {
	pod, err := test.FixturePod(".", "pod-injection-enabled-00.json")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	sidecar, err := test.FixtureContainer(".", "sidecar-container.json")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	podPatch := NewPodPatch(pod)
	podPatch.addContainerPatch(sidecar)
	podPatch.addAnnotationPatch()

	expectedOps := []*patchOp{
		&patchOp{Op: "add", Path: patchPathContainer, Value: sidecar},
		&patchOp{Op: "add", Path: patchPathAnnotation, Value: map[string]string{annotationKeySidecarInjection: "false"}},
	}
	expected, err := json.Marshal(expectedOps)
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	actual, err := json.Marshal(podPatch.patchOps)
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Content mismatch\nExpected: %s\nActual: %s", expected, actual)
	}
}
