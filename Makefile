OUTDIR = bin
APPNAME = mailer
CONFIG = config.toml

build:
	CGO_ENABLED=0 go build -o ${OUTDIR}/${APPNAME} .

build-for-docker:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTDIR}/${APPNAME}-linux .

doc:
	godoc -http=:6060 -index

vet:
	go vet ./...

run: build
	./$(OUTDIR)/$(APPNAME) -config $(CONFIG)

test:
	go test ./...

VERSION = 2407-amd64
docker-build: build-for-docker
	docker build \
	  --platform linux/amd64 \
	  --progress plain \
	  -t schmooser/go-mailer:$(VERSION) .


docker-push:
	docker push schmooser/go-mailer:$(VERSION)
