# source: http://zduck.com/2014/go-project-structure-and-dependencies/
.PHONY: all build doc fmt lint run test vendor_clean vendor_get vendor_update vet

# Prepend our vendor directory to the system GOPATH
# so that import path resolution will prioritize
# our third party snapshots.

all: vendor_get build

build:
	CGO_ENABLED=0 GOOS=linux \
	go build -a -installsuffix cgo -v -o ${OUTDIR}/${APPNAME} .

doc:
	godoc -http=:6060 -index

# http://golang.org/cmd/go/#hdr-Run_gofmt_on_package_sources
fmt:
	go fmt ./...

# https://github.com/golang/lint
# go get github.com/golang/lint/golint
lint:
	golint ./

run: build
	./${OUTDIR}/${APPNAME}

test:
	go test ./...

vendor_clean:
	rm -dRf ./vendor

# We have to set GOPATH to just the vendor
# directory to ensure that `go get` doesn't
# update packages in our primary GOPATH instead.
# This will happen if you already have the package
# installed in GOPATH since `go get` will use
# that existing location as the destination.
vendor_get: vendor_clean
	GOPATH=${PWD}/vendor go get -d -v \
	github.com/BurntSushi/toml \
	github.com/labstack/echo/... \
	github.com/mostafah/mandrill \
	gopkg.in/inconshreveable/log15.v2 \
	github.com/schmooser/go-echolog15 \
	&& mv ${PWD}/vendor/src/* ${PWD}/vendor \
	&& rmdir ${PWD}/vendor/src

vendor_update: vendor_get
	rm -rf `find ./vendor -type d -name .git` \
	&& rm -rf `find ./vendor -type d -name .hg`

# http://godoc.org/code.google.com/p/go.tools/cmd/vet
# go get code.google.com/p/go.tools/cmd/vet
vet:
	mv ${PWD}/vendor ${PWD}/_vendor \
	&& go vet ./... \
	&& mv ${PWD}/_vendor ${PWD}/vendor
