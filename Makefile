VERSION = 0.0.1

build:
	docker build --rm \
		--build-arg BUILD_DATE="`date +'%Y-%m-%d %T %z'`" \
		--build-arg VCS_REF=`git rev-parse --short HEAD` \
		--build-arg VCS_URL="https://github.com/ihcsim/admission-webhook" \
		--build-arg VERSION="$(VERSION)" \
		-t isim/admission-webhook .
