package test

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

// FixtureHTTPRequestBody returns the content of the specified file as a slice of bytes.
// If the file doesn't exist in the 'test/data' folder, an error will be returned.
func FixtureHTTPRequestBody(filename, prefix string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(prefix, "test", "data", filename))
}

// FixtureAdmissionReview returns the content of the specified file as an AdmissionReview type. An error will be returned if:
// i. the file doesn't exist in the 'test/data' folder or,
// ii. the file content isn't a valid JSON structure that can be unmarshalled into AdmissionReview type
func FixtureAdmissionReview(filename, prefix string) (*admissionv1beta1.AdmissionReview, error) {
	b, err := ioutil.ReadFile(filepath.Join(prefix, "test", "data", filename))
	if err != nil {
		return nil, err
	}
	var admissionReview admissionv1beta1.AdmissionReview
	if err := json.Unmarshal(b, &admissionReview); err != nil {
		return nil, err
	}

	return &admissionReview, nil
}

// FixtureAdmissionResponse returns the content of the specified file as an AdmissionResponse type. An error will be returned if:
// i. the file doesn't exist in the 'test/data' folder or
// ii. the file content isn't a valid JSON structure that can be unmarshalled into AdmissionResponse type
func FixtureAdmissionResponse(prefix string) (*admissionv1beta1.AdmissionResponse, error) {
	b, err := ioutil.ReadFile(filepath.Join(prefix, "test", "data", "admission-response.json"))
	if err != nil {
		return nil, err
	}
	var admissionResponse admissionv1beta1.AdmissionResponse
	if err := json.Unmarshal(b, &admissionResponse); err != nil {
		return nil, err
	}

	return &admissionResponse, nil
}

// FixturePod returns the content of the specified file as an Pod type. An error will be returned if:
// i. the file doesn't exist in the 'test/data' folder or
// ii. the file content isn't a valid JSON structure that can be unmarshalled into Pod type
func FixturePod(prefix, filename string) (*corev1.Pod, error) {
	b, err := ioutil.ReadFile(filepath.Join(prefix, "test", "data", filename))
	if err != nil {
		return nil, err
	}

	var pod corev1.Pod
	if err := json.Unmarshal(b, &pod); err != nil {
		return nil, err
	}

	return &pod, nil
}

// FixtureContainer returns the content of the specified file as a Container type. An error will be returned if:
// i. the file doesn't exist in the 'test/data' folder or
// ii. the file content isn't a valid JSON structure that can be unmarshalled into Cnotainer type
func FixtureContainer(prefix, filename string) (*corev1.Container, error) {
	b, err := ioutil.ReadFile(filepath.Join(prefix, "test", "data", filename))
	if err != nil {
		return nil, err
	}

	var container corev1.Container
	if err := json.Unmarshal(b, &container); err != nil {
		return nil, err
	}

	return &container, nil
}
