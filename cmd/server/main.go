package main

import (
	"flag"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	port     = ""
	certFile = ""
	keyFile  = ""
	debug    = ""

	log = logrus.New()
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

	s, err := NewWebhookServer(port, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	s.Handler = http.HandlerFunc(s.serve)

	if err := s.ListenAndServeTLS("", ""); err != nil {
		log.Fatal(err)
	}
}
