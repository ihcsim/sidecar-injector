package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"

	webhook "github.com/ihcsim/admission-webhook"
	"github.com/sirupsen/logrus"
)

// WebhookServer is the webhook's TLS server. Its embedded http.Server handles all incoming requests. The webhook performs the mutation and interacts with the k8s API Server.
type WebhookServer struct {
	*http.Server
	*webhook.Webhook
	*logrus.Entry
}

// NewWebhookServer returns a new instance of the WebhookServer.
func NewWebhookServer(port, certFile, keyFile string) (*WebhookServer, error) {
	c, err := tlsConfig(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	server := &http.Server{
		Addr:      ":" + port,
		TLSConfig: c,
	}

	webhook, err := webhook.New()
	if err != nil {
		return nil, err
	}
	webhook.SetLogLevel(log.Level)
	requestLogger := logrus.NewEntry(log)

	return &WebhookServer{server, webhook, requestLogger}, nil
}

func (w *WebhookServer) serve(res http.ResponseWriter, req *http.Request) {
	w.Data = logrus.Fields{"Remote Addr": req.RemoteAddr}

	var (
		data []byte
		err  error
	)
	if req.Body != nil {
		data, err = ioutil.ReadAll(req.Body)
		if err != nil {
			w.handleRequestError(res, err, http.StatusBadRequest)
			return
		}
		w.Debugf("HTTP Request body: %s", data)
	}

	if len(data) == 0 {
		return
	}

	response := w.Mutate(data)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		w.handleRequestError(res, err, http.StatusInternalServerError)
		return
	}
	w.Debugf("HTTP Response body: %s", responseJSON)

	if _, err := res.Write(responseJSON); err != nil {
		w.handleRequestError(res, err, http.StatusInternalServerError)
		return
	}
}

func (w *WebhookServer) handleRequestError(res http.ResponseWriter, err error, code int) {
	w.WithFields(logrus.Fields{
		"code": code,
	}).Error(err)
	http.Error(res, err.Error(), code)
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
