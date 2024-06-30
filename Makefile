OUTDIR = bin
APPNAME = mailer
CONFIG = config.toml

build:
	CGO_ENABLED=0 go build -o ${OUTDIR}/${APPNAME} .

doc:
	godoc -http=:6060 -index

vet:
	go vet ./...

run: build
	./$(OUTDIR)/$(APPNAME) -config $(CONFIG)

test:
	go test ./...
