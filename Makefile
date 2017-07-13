VERSION := $(shell cat VERSION)
OS=$(shell lsb_release -si)
ARCH=$(shell uname -m | sed 's/x86_//;s/i[3-6]86/32/')

deps:
	go get -t -d -v ./...

deps-release:

build:
	mkdir -p dist/
	go build -o dist/nxircd_${VERSION}_linux_${ARCH} .
	GOARCH=386 GOOS=linux go build -o dist/nxircd_${VERSION}_linux_386 .

dist: build
	ghr -t ${GITHUB_TOKEN} -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} ${VERSION} dist/

test:
	go vet ./...
	go test -v -race -cover ./...

clean:
	rm -frv dist
