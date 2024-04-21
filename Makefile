GOARCH:="amd64"
GOOS:="linux"
CGO_ENABLED:=1

all: lint test build

clean:
	rm -r ./out/

lint:
	echo "TODO"

test:
	go test -v ./...

build: ./out/dhcpdebug

./out:
	mkdir -p out

./out/dhcpdebug:
	GOARCH=${GOARCH} GOOS=${GOOS} CGO_ENABLED=${CGO_ENABLED} go build -o ./out/dhcpdebug ./cmd/main.go
