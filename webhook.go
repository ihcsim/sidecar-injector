package webhook

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	errNilAdmissionReviewInput = fmt.Errorf("AdmissionReview input object can't be nil")
)

// Webhook is a Kubernetes mutating admission webhook that mutates pods admission requests by injecting sidecar container spec into the pod spec.
type Webhook struct {
	log          *logrus.Logger
	deserializer runtime.Decoder
}

// New returns a new instance of Webhook
func New() *Webhook {
	var (
		scheme = runtime.NewScheme()
		codecs = serializer.NewCodecFactory(scheme)
	)
	return &Webhook{
		log:          logrus.New(),
		deserializer: codecs.UniversalDeserializer(),
	}
}

// Mutate changes the pod spec defined in data by injecting sidecar container spec into the spec. The admission review object returns contains the original request and the response with the mutated pod spec.
func (w *Webhook) Mutate(data []byte) *admissionv1beta1.AdmissionReview {
	admissionReview, err := w.decode(data)
	if err != nil {
		w.log.Info("Failed to decode data. Reason: ", err)
		admissionReview.Response = &admissionv1beta1.AdmissionResponse{
			UID: admissionReview.Request.UID,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
		return admissionReview
	}

	admissionResponse, err := w.inject(admissionReview)
	if err != nil {
		admissionReview.Response = &admissionv1beta1.AdmissionResponse{
			UID: admissionReview.Request.UID,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
		return admissionReview
	}
	admissionReview.Response = admissionResponse
	admissionReview.Response.UID = admissionReview.Request.UID

	requestJSON, _ := json.Marshal(admissionReview.Request)
	w.log.Debugf("Admission request: %s", requestJSON)

	responseJSON, _ := json.Marshal(admissionReview.Response)
	w.log.Debugf("Admission response: %s", responseJSON)

	return admissionReview
}

func (w *Webhook) decode(data []byte) (*admissionv1beta1.AdmissionReview, error) {
	admissionReview := admissionv1beta1.AdmissionReview{}
	_, _, err := w.deserializer.Decode(data, nil, &admissionReview)
	return &admissionReview, err
}

func (w *Webhook) inject(ar *admissionv1beta1.AdmissionReview) (*admissionv1beta1.AdmissionResponse, error) {
	if ar == nil {
		return nil, errNilAdmissionReviewInput
	}

	request := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		return nil, err
	}

	var (
		patch     = []byte(`Hello World`)
		patchType = admissionv1beta1.PatchTypeJSONPatch
	)
	admissionResponse := &admissionv1beta1.AdmissionResponse{
		Allowed:   true,
		Patch:     patch,
		PatchType: &patchType,
	}

	return admissionResponse, nil
}
