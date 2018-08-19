package webhook

import (
	"reflect"
	"testing"

	"github.com/ihcsim/admission-webhook/test"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMutate(t *testing.T) {
	data, err := test.FixtureHTTPRequestBody("http-request-body-valid.json", ".")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	expected, err := test.FixtureAdmissionReview("admission-review-request-response.json", ".")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	webhook := New()
	actual := webhook.Mutate(data)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Content mismatch\nExpected: %+v\nActual: %+v", expected, actual)
	}
}

func TestDecode(t *testing.T) {
	webhook := New()

	t.Run("With Nil Input", func(t *testing.T) {
		actual, err := webhook.decode(nil)
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
		in, err := test.FixtureHTTPRequestBody("http-request-body-valid.json", ".")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		expected, err := test.FixtureAdmissionReview("admission-review-request-only.json", ".")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		actual, err := webhook.decode(in)
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Decoded content mismatch\nExpected: %+v\nActual: %+v", expected, actual)
		}
	})

	t.Run("With Invalid HTTP Request Body", func(t *testing.T) {
		in, err := test.FixtureHTTPRequestBody("http-request-body-invalid.json", ".")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if _, err := webhook.decode(in); err == nil {
			t.Error("Expected test to fail with malformed JSON error")
		}
	})
}

func TestInject(t *testing.T) {
	webhook := New()

	t.Run("With Nil input", func(t *testing.T) {
		_, err := webhook.inject(nil)
		if err == nil {
			t.Error("Expected error didn't occur")
		}

		if !reflect.DeepEqual(err, errNilAdmissionReviewInput) {
			t.Errorf("Mismatch returned error.\nExpected: %q\nActual: %q", errNilAdmissionReviewInput, err)
		}
	})

	t.Run("With Valid Admission Review", func(t *testing.T) {
		admissionReview, err := test.FixtureAdmissionReview("admission-review-request-only.json", ".")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		expected, err := test.FixtureAdmissionResponse(".")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		actual, err := webhook.inject(admissionReview)
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Mismatch content\nExpected: %+v\nActual: %+v", expected, actual)
		}
	})
}
