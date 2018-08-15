VERSION ?= 0.0.1

image: image/build image/push

image/build:
	docker build --rm \
		--build-arg BUILD_DATE="`date +'%Y-%m-%d %T %z'`" \
		--build-arg VCS_REF=`git rev-parse --short HEAD` \
		--build-arg VCS_URL="https://github.com/ihcsim/admission-webhook" \
		--build-arg VERSION="$(VERSION)" \
		-t isim/admission-webhook:$(VERSION) .

image/push:
	docker push isim/admission-webhook:$(VERSION)

container/run:
	docker run -d --name admission-webhook -v `pwd`/tls/server:/etc/secret isim/admission-webhook:$(VERSION)

container/clean:
	docker stop admission-webhook
	docker rm admission-webhook

tls/ca: tls/ca/key tls/ca/cert

tls/dir:
	mkdir -p tls

tls/ca/key: tls/dir
	openssl genrsa -out tls/ca.key 4096

tls/ca/cert: tls/dir
	mkdir -p tls
	openssl req -x509 -new -nodes -key tls/ca.key -sha256 -days 365 -out tls/ca.crt

tls/server: tls/server/key tls/server/csr tls/server/cert

tls/server/dir:
	mkdir -p tls/server

tls/server/key: tls/server/dir
	openssl genrsa -out tls/server/server.key 2048

tls/server/csr: tls/server/dir
	openssl req -new -key tls/server/server.key -out tls/server/server.csr

tls/server/cert: tls/server/dir
	openssl x509 -req -in tls/server/server.csr -CA tls/ca.crt -CAkey tls/ca.key -CAcreateserial -out tls/server/server.crt -days 365 -sha256
