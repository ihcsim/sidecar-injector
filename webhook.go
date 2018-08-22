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
	"k8s.io/client-go/kubernetes"
)

const (
	annotationKeySidecarInjection = "sidecar.example.org/inject"
	configMapSidecar              = "sidecar-spec"
	defaultNamespace              = "default"
)

var (
	errNilAdmissionReviewInput = fmt.Errorf("AdmissionReview input object can't be nil")

	NewClient = NewClientset
)

// Webhook is a Kubernetes mutating admission webhook that mutates pods admission requests by injecting sidecar container spec into the pod spec.
type Webhook struct {
	logger       *logrus.Logger
	deserializer runtime.Decoder
	Client       kubernetes.Interface
}

// New returns a new instance of Webhook.
func New() (*Webhook, error) {
	var (
		scheme = runtime.NewScheme()
		codecs = serializer.NewCodecFactory(scheme)
		logger = logrus.New()
	)

	client, err := NewClient()
	if err != nil {
		return nil, err
	}

	return &Webhook{
		logger:       logger,
		deserializer: codecs.UniversalDeserializer(),
		Client:       client,
	}, nil
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

	opt := metav1.GetOptions{}
	sidecar, err := w.sidecarFromConfigMap(configMapSidecar, defaultNamespace, opt)
	if err != nil {
		return nil, err
	}
	w.logger.Debugf("Sidecar: %+v", sidecar)

	podPatch := NewPodPatch(&pod)
	podPatch.addContainerPatch(sidecar)
	podPatch.addAnnotationPatch()

	patchJSON, err := json.Marshal(podPatch.patchOps)
	if err != nil {
		return nil, err
	}

	patchType := admissionv1beta1.PatchTypeJSONPatch
	admissionResponse := &admissionv1beta1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchJSON,
		PatchType: &patchType,
	}

	return admissionResponse, nil
}

func (w *Webhook) ignore(pod *corev1.Pod) bool {
	annotations := pod.ObjectMeta.GetAnnotations()
	inject, err := strconv.ParseBool(annotations[annotationKeySidecarInjection])
	if err != nil {
		// annontation isn't specified so return false to not ignore this pod
		return false
	}

	return !inject
}

func (w *Webhook) sidecarFromConfigMap(name, namespace string, opt metav1.GetOptions) (*corev1.Container, error) {
	configMap, err := w.Client.CoreV1().ConfigMaps(namespace).Get(name, opt)
	if err != nil {
		return nil, err
	}

	var c corev1.Container
	if err := json.Unmarshal([]byte(configMap.Data["sidecar.json"]), &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// SetLogLevel sets the log level of the webhook's logger.
func (w *Webhook) SetLogLevel(level logrus.Level) {
	w.logger.SetLevel(level)
}
