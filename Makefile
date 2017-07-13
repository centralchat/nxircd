VERSION := $(shell cat VERSION)

deps:
	go get -t -d -v ./...

deps-release:

build:
	mkdir -p dist/
	go build -o dist/nxircd_${VERSION}_{{.OS}}_{{.Arch}} .

dist: build
	ghr -t ${GITHUB_TOKEN} -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} ${VERSION} dist/

test:
	go vet ./...
	go test -v -race -cover ./...

clean:
	rm -frv dist
