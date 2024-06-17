build:
	CGO_ENABLED=0 GOOS=linux \
	go build -a -v -o ${OUTDIR}/${APPNAME} .

doc:
	godoc -http=:6060 -index

vet:
	go vet ./...

run: build
	./${OUTDIR}/${APPNAME}

test:
	go test ./...
