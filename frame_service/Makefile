APP?=frameservice
PROJECT?=github.com/enzo1920/frame_service/frame_service

RELEASE?=0.0.1
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')


GOOS?=linux
GOARCH?=amd64


clean:
			rm -f ${APP}
build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o ${APP}
run: build
			./${APP}
test:
			go test -v -race ./...