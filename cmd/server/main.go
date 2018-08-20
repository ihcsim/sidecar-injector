package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	webhook "github.com/ihcsim/admission-webhook"
	"github.com/sirupsen/logrus"
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
		requestLogger.Debugf("HTTP Request body: %s", data)
	}

	if len(data) == 0 {
		return
	}

	webhook := webhook.New()
	webhook.SetLogLevel(log.Level)
	response := webhook.Mutate(data)
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
