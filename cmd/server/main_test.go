package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMutate(t *testing.T) {
	data, err := ioutil.ReadFile(filepath.Join("test-data", "http-request-body-valid"))
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	expected, err := fixtureAdmissionReview("admission-review-request-response.json")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	actual := mutate(data)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Content mismatch\nExpected: %+v\nActual: %+v", expected, actual)
	}
}

func TestDecode(t *testing.T) {
	t.Run("With Nil Input", func(t *testing.T) {
		actual, err := decode(nil)
		if err != nil {
			t.Fatal("Unexpected error", err)
		}

		expected := &admissionv1beta1.AdmissionReview{
			metav1.TypeMeta{},
			nil,
			nil,
		}
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Decoded content mismatch\nExpected: %+v\nActual: %+v", expected, actual)
		}
	})

	t.Run("With Valid HTTP Request Body", func(t *testing.T) {
		in, err := ioutil.ReadFile(filepath.Join("test-data", "http-request-body-valid"))
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		expected, err := fixtureAdmissionReview("admission-review-request-only.json")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		actual, err := decode(in)
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Decoded content mismatch\nExpected: %+v\nActual: %+v", expected, actual)
		}
	})

	t.Run("With Invalid HTTP Request Body", func(t *testing.T) {
		in, err := ioutil.ReadFile(filepath.Join("test-data", "http-request-body-invalid"))
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if _, err := decode(in); err == nil {
			t.Error("Expected test to fail with malformed JSON error")
		}
	})
}

func TestInject(t *testing.T) {
	t.Run("With Nil input", func(t *testing.T) {
		_, err := inject(nil)
		if err == nil {
			t.Error("Expected error didn't occur")
		}

		if !reflect.DeepEqual(err, errNilAdmissionReviewInput) {
			t.Errorf("Mismatch returned error.\nExpected: %q\nActual: %q", errNilAdmissionReviewInput, err)
		}
	})

	t.Run("With Valid Admission Review", func(t *testing.T) {
		admissionReview, err := fixtureAdmissionReview("admission-review-request-only.json")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		expected, err := fixtureAdmissionResponse()
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		actual, err := inject(admissionReview)
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Mismatch content\nExpected: %+v\nActual: %+v", expected, actual)
		}
	})
}

func TestServe(t *testing.T) {
	t.Run("With Empty HTTP Request Body", func(t *testing.T) {
		in := bytes.NewReader(nil)
		request := httptest.NewRequest(http.MethodGet, "/", in)

		recorder := httptest.NewRecorder()
		serve(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusOK, recorder.Code)
		}

		if reflect.DeepEqual(recorder.Body.Bytes(), []byte("")) {
			t.Errorf("Content mismatch. Expected HTTP response body to be empty %v", recorder.Body.Bytes())
		}
	})

	t.Run("With Valid HTTP Request Body", func(t *testing.T) {
		body, err := ioutil.ReadFile(filepath.Join("test-data", "http-request-body-valid"))
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		in := bytes.NewReader(body)
		request := httptest.NewRequest(http.MethodGet, "/", in)

		recorder := httptest.NewRecorder()
		serve(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusOK, recorder.Code)
		}

		expected, err := fixtureAdmissionReview("admission-review-request-response.json")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		var actual admissionv1beta1.AdmissionReview
		if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if !reflect.DeepEqual(actual, *expected) {
			t.Errorf("Content mismatch\nExpected: %+v\nActual: %+v", *expected, actual)
		}
	})
}

func TestHandleRequestError(t *testing.T) {
	var (
		errMsg   = "Some test error"
		recorder = httptest.NewRecorder()
		err      = fmt.Errorf(errMsg)
		logger   = logrus.New()
	)

	logger.SetOutput(ioutil.Discard)
	requestLogger := logrus.NewEntry(logger)

	handleRequestError(recorder, err, http.StatusInternalServerError, requestLogger)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusInternalServerError, recorder.Code)
	}

	if strings.TrimSpace(recorder.Body.String()) != errMsg {
		t.Errorf("HTTP response body mismatch. Expected: %q. Actual: %q", errMsg, recorder.Body.String())
	}
}

func fixtureAdmissionReview(filename string) (*admissionv1beta1.AdmissionReview, error) {
	b, err := ioutil.ReadFile(filepath.Join("test-data", filename))
	if err != nil {
		return nil, err
	}
	var expected admissionv1beta1.AdmissionReview
	if err := json.Unmarshal(b, &expected); err != nil {
		return nil, err
	}

	return &expected, nil
}

func fixtureAdmissionResponse() (*admissionv1beta1.AdmissionResponse, error) {
	b, err := ioutil.ReadFile(filepath.Join("test-data", "admission-response.json"))
	if err != nil {
		return nil, err
	}
	var expected admissionv1beta1.AdmissionResponse
	if err := json.Unmarshal(b, &expected); err != nil {
		return nil, err
	}

	return &expected, nil
}
