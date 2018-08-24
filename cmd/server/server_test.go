package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	webhook "github.com/ihcsim/sidecar-injector"
	"github.com/ihcsim/sidecar-injector/test"
	"github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
)

var testServer *WebhookServer

func TestMain(m *testing.M) {
	// mock out the k8s clientset constructor
	webhook.NewClient = test.NewFakeClient

	// create a webhook which uses its fake client to seed the sidecar configmap
	w, err := initWebhookWithConfigMap()
	if err != nil {
		panic(err)
	}

	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	logger := logrus.NewEntry(log)
	testServer = &WebhookServer{nil, w, logger}

	os.Exit(m.Run())
}

func TestServe(t *testing.T) {
	t.Run("With Empty HTTP Request Body", func(t *testing.T) {
		in := bytes.NewReader(nil)
		request := httptest.NewRequest(http.MethodGet, "/", in)

		recorder := httptest.NewRecorder()
		testServer.serve(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusOK, recorder.Code)
		}

		if reflect.DeepEqual(recorder.Body.Bytes(), []byte("")) {
			t.Errorf("Content mismatch. Expected HTTP response body to be empty %v", recorder.Body.Bytes())
		}
	})

	t.Run("With Valid HTTP Request Body", func(t *testing.T) {
		body, err := test.FixtureHTTPRequestBody("http-request-body-valid.json", "../..")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		in := bytes.NewReader(body)
		request := httptest.NewRequest(http.MethodGet, "/", in)

		recorder := httptest.NewRecorder()
		testServer.serve(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusOK, recorder.Code)
		}

		expected, err := test.FixtureAdmissionReview("admission-review-request-response.json", "../..")
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

	t.Run("With Valid HTTP Request Body (ignore pod)", func(t *testing.T) {
		body, err := test.FixtureHTTPRequestBody("http-request-body-valid-ignore-pod.json", "../..")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		in := bytes.NewReader(body)
		request := httptest.NewRequest(http.MethodGet, "/", in)

		recorder := httptest.NewRecorder()
		testServer.serve(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusOK, recorder.Code)
		}

		expected, err := test.FixtureAdmissionReview("admission-review-request-response-ignore-pod.json", "../..")
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		var actual admissionv1beta1.AdmissionReview
		if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if !reflect.DeepEqual(actual, *expected) {
			t.Errorf("Content mismatch\nExpected: %s\nActual: %s", expected, actual)
		}
	})
}

func TestHandleRequestError(t *testing.T) {
	var (
		errMsg   = "Some test error"
		recorder = httptest.NewRecorder()
		err      = fmt.Errorf(errMsg)
	)

	testServer.handleRequestError(recorder, err, http.StatusInternalServerError)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("HTTP response status mismatch. Expected: %d. Actual: %d", http.StatusInternalServerError, recorder.Code)
	}

	if strings.TrimSpace(recorder.Body.String()) != errMsg {
		t.Errorf("HTTP response body mismatch. Expected: %q. Actual: %q", errMsg, recorder.Body.String())
	}
}

func TestNewWebhookServer(t *testing.T) {
	// sample cert and key pem files from https://golang.org/src/crypto/tls/tls_test.go
	var (
		rsaCertPEM = `-----BEGIN CERTIFICATE-----
MIIB0zCCAX2gAwIBAgIJAI/M7BYjwB+uMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTIwOTEyMjE1MjAyWhcNMTUwOTEyMjE1MjAyWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBANLJ
hPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wok/4xIA+ui35/MmNa
rtNuC+BdZ1tMuVCPFZcCAwEAAaNQME4wHQYDVR0OBBYEFJvKs8RfJaXTH08W+SGv
zQyKn0H8MB8GA1UdIwQYMBaAFJvKs8RfJaXTH08W+SGvzQyKn0H8MAwGA1UdEwQF
MAMBAf8wDQYJKoZIhvcNAQEFBQADQQBJlffJHybjDGxRMqaRmDhX0+6v02TUKZsW
r5QuVbpQhH6u+0UgcW0jp9QwpxoPTLTWGXEWBBBurxFwiCBhkQ+V
-----END CERTIFICATE-----
`

		rsaKeyPEM = `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBANLJhPHhITqQbPklG3ibCVxwGMRfp/v4XqhfdQHdcVfHap6NQ5Wo
k/4xIA+ui35/MmNartNuC+BdZ1tMuVCPFZcCAwEAAQJAEJ2N+zsR0Xn8/Q6twa4G
6OB1M1WO+k+ztnX/1SvNeWu8D6GImtupLTYgjZcHufykj09jiHmjHx8u8ZZB/o1N
MQIhAPW+eyZo7ay3lMz1V01WVjNKK9QSn1MJlb06h/LuYv9FAiEA25WPedKgVyCW
SmUwbPw8fnTcpqDWE3yTO3vKcebqMSsCIBF3UmVue8YU3jybC3NxuXq3wNm34R8T
xVLHwDXh/6NJAiEAl2oHGGLz64BuAfjKrqwz7qMYr9HCLIe/YsoWq/olzScCIQDi
D2lWusoe2/nEqfDVVWGWlyJ7yOmqaVm/iNUN9B2N2g==
-----END RSA PRIVATE KEY-----`
	)

	certFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	defer os.Remove(certFile.Name())

	if err := ioutil.WriteFile(certFile.Name(), []byte(rsaCertPEM), 0); err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	keyFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}
	defer os.Remove(certFile.Name())

	if err := ioutil.WriteFile(keyFile.Name(), []byte(rsaKeyPEM), 0); err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	port := "7070"
	server, err := NewWebhookServer(port, certFile.Name(), keyFile.Name())
	if err != nil {
		t.Fatal("Unexpected error: ", err)
	}

	if server.Addr != fmt.Sprintf(":%s", port) {
		t.Errorf("Expected server address to be :%q", port)
	}
}

func initWebhookWithConfigMap() (*webhook.Webhook, error) {
	fixture, err := webhook.New()
	if err != nil {
		return nil, err
	}

	// seed the sidecar configmap with the fake client
	configMap, err := test.FixtureConfigMap("../..", "sidecar-configmap.json")
	if err != nil {
		return nil, err
	}

	if _, err := fixture.Client.CoreV1().ConfigMaps(test.DefaultNamespace).Create(configMap); err != nil {
		return nil, err
	}

	return fixture, nil
}
