VERSION := $(shell cat VERSION)
OS=$(shell lsb_release -si)
ARCH=$(shell uname -m | sed 's/x86_//;s/i[3-6]86/32/')

deps:
	go get -t -d -v ./...

deps-release:

build:
	mkdir -p dist/
	GOOS=linux GOARCH=amd64 go build -o dist/nxircd_${VERSION}_linux_amd64 .
	GOOS=linux GOARCH=386 go build -o dist/nxircd_${VERSION}_linux_x86 .
	GOOS=darwin GOARCH=amd64 go build -o dist/nxircd_${VERSION}_osx_amd64 .
	GOOS=darwin GOARCH=386 go build -o dist/nxircd_${VERSION}_osx_x86 .
	GOOS=windows GOARCH=amd64 go build -o dist/nxircd_${VERSION}_windows_64
    GOOS=windows GOARCH=386  build -o dist/nxircd_${VERSION}_windows_x86

dist: build
	ghr -t ${GITHUB_TOKEN} -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} ${VERSION} dist/

test:
	go vet ./...
	go test -v -race -cover ./...

clean:
	rm -frv dist
