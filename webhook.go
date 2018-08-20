package webhook

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	logger       *logrus.Logger
	deserializer runtime.Decoder
}

// New returns a new instance of Webhook
func New() *Webhook {
	var (
		scheme = runtime.NewScheme()
		codecs = serializer.NewCodecFactory(scheme)
		logger = logrus.New()
	)

	return &Webhook{
		logger:       logger,
		deserializer: codecs.UniversalDeserializer(),
	}
}

// Mutate changes the pod spec defined in data by injecting sidecar container spec into the spec. The admission review object returns contains the original request and the response with the mutated pod spec.
func (w *Webhook) Mutate(data []byte) *admissionv1beta1.AdmissionReview {
	admissionReview, err := w.decode(data)
	if err != nil {
		w.logger.Info("Failed to decode data. Reason: ", err)
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
	w.logger.Debugf("Admission request: %s", requestJSON)

	responseJSON, _ := json.Marshal(admissionReview.Response)
	w.logger.Debugf("Admission response: %s", responseJSON)

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
	w.logger.Debugf("Request JSON object: %s", request.Object.Raw)

	var pod corev1.Pod
	if err := json.Unmarshal(request.Object.Raw, &pod); err != nil {
		return nil, err
	}
	w.logger.Debugf("Pod: %+v", pod)

	if w.ignore(&pod) {
		return &admissionv1beta1.AdmissionResponse{
			Allowed: true,
		}, nil
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

func (w *Webhook) ignore(pod *corev1.Pod) bool {
	annotations := pod.ObjectMeta.GetAnnotations()
	inject, err := strconv.ParseBool(annotations["sidecar.example.org/inject"])
	if err != nil {
		// if annontation isn't specified, don't ignore pod
		return false
	}

	return !inject
}

// SetLogLevel sets the log level of the webhook's logger.
func (w *Webhook) SetLogLevel(level logrus.Level) {
	w.logger.SetLevel(level)
}
