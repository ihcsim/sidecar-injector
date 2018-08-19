VERSION ?= 0.0.1
DEBUG_ENABLED ?= false

test:
	cd cmd/server
	go test -v -cover -race ./...

build:
	docker build --rm \
		--build-arg BUILD_DATE="`date +'%Y-%m-%d %T %z'`" \
		--build-arg VCS_REF=`git rev-parse --short HEAD` \
		--build-arg VCS_URL="https://github.com/ihcsim/admission-webhook" \
		--build-arg VERSION="$(VERSION)" \
		-t isim/admission-webhook:$(VERSION) .

push:
	docker push isim/admission-webhook:$(VERSION)

tls/ca:
	rm -rf tls/ca
	mkdir -p tls/ca
	openssl genrsa -out tls/ca/ca.key 4096
	openssl req -x509 -new -nodes -key tls/ca/ca.key -sha256 -days 365 -out tls/ca/ca.crt

tls/server:
	rm -rf tls/server
	mkdir -p tls/server
	openssl genrsa -out tls/server/server.key 2048
	openssl req -new -key tls/server/server.key -out tls/server/server.csr -config tls/san.cnf
	openssl x509 -req -in tls/server/server.csr -CA tls/ca/ca.crt -CAkey tls/ca/ca.key -CAcreateserial -out tls/server/server.crt -days 365 -sha256 -extensions req_ext -extfile tls/san.cnf

CA_BUNDLE=$(shell cat tls/ca/ca.crt | base64 -w 0)
TLS_CERT=$(shell cat tls/server/server.crt | base64 -w 0)
TLS_KEY=$(shell cat tls/server/server.key | base64 -w 0)

deploy:
	sed -e s/\$$\{CA_BUNDLE\}/"$(CA_BUNDLE)"/ -e s/\$$\{TLS_CERT\}/"$(TLS_CERT)"/ -e s/\$$\{TLS_KEY\}/"$(TLS_KEY)"/ -e s/\$$\{DEBUG_ENABLED\}/${DEBUG_ENABLED}/ charts/deployment.yaml | kubectl apply -f -

purge:
	kubectl delete all -l app=admission-webhook
