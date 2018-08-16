package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	port     = ""
	certFile = ""
	keyFile  = ""
	debug    = ""

	log = logrus.New()

	server = tlsServer
)

func init() {
	flag.StringVar(&port, "port", "443", "Port that this webhook admission server listens on")
	flag.StringVar(&certFile, "cert-file", "/etc/secret/tls.crt", "Location of the TLS cert file")
	flag.StringVar(&keyFile, "key-file", "/etc/secret/tls.key", "Location of the TLS private key file")
	flag.StringVar(&debug, "debug", "false", "Set to 'true' to enable more verbose debug mode")
	flag.Parse()

	if strings.ToLower(debug) == "true" {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("Starting in debug mode...")
	} else {
		log.SetLevel(logrus.InfoLevel)
	}
	log.SetOutput(os.Stdout)
}

func main() {
	log.Infof("Listening at port %s... ", port)
	log.Infof("Using TLS cert at %s and key at %s...", certFile, keyFile)

	s, err := server(port, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := s.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}

func serve(res http.ResponseWriter, req *http.Request) {
	requestLogger := logrus.NewEntry(log)
	requestLogger.Data = logrus.Fields{"Remote Addr": req.RemoteAddr}

	var (
		data []byte
		err  error
	)
	if req.Body != nil {
		data, err = ioutil.ReadAll(req.Body)
		if err != nil {
			handleRequestError(res, err, http.StatusBadRequest, requestLogger)
			return
		}
	}
	requestLogger.Debugf("HTTP Request body: %s", data)

	response := mutate(data)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		handleRequestError(res, err, http.StatusInternalServerError, requestLogger)
		return
	}
	requestLogger.Debugf("HTTP Response body: %s", responseJSON)

	if _, err := res.Write(responseJSON); err != nil {
		handleRequestError(res, err, http.StatusInternalServerError, requestLogger)
		return
	}
}

func mutate(data []byte) *admissionv1beta1.AdmissionReview {
	admissionReview, err := decode(data)
	if err != nil {
		log.Info("Failed to decode data. Reason: ", err)
		admissionReview.Response = &admissionv1beta1.AdmissionResponse{
			UID: admissionReview.Request.UID,
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
		return admissionReview
	}

	admissionResponse, err := inject(admissionReview)
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
	log.Debugf("Admission request: %s", requestJSON)

	responseJSON, _ := json.Marshal(admissionReview.Response)
	log.Debugf("Admission response: %s", responseJSON)

	return admissionReview
}

func decode(data []byte) (*admissionv1beta1.AdmissionReview, error) {
	var (
		admissionReview = admissionv1beta1.AdmissionReview{}
		scheme          = runtime.NewScheme()
		codecs          = serializer.NewCodecFactory(scheme)
		deserializer    = codecs.UniversalDeserializer()
	)

	if _, _, err := deserializer.Decode(data, nil, &admissionReview); err != nil {
		return &admissionReview, err
	}

	return &admissionReview, nil
}

func inject(ar *admissionv1beta1.AdmissionReview) (*admissionv1beta1.AdmissionResponse, error) {
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

func tlsServer(port, certFile, keyFile string) (*http.Server, error) {
	c, err := tlsConfig(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:      ":" + port,
		Handler:   http.HandlerFunc(serve),
		TLSConfig: c,
	}, nil
}

func tlsConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

func handleRequestError(w http.ResponseWriter, err error, code int, requestLogger *logrus.Entry) {
	requestLogger.WithFields(logrus.Fields{
		"code": code,
	}).Error(err)
	http.Error(w, err.Error(), code)
}
