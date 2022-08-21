FROM golang:1.18-bullseye

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

ENV APP_HOME /go/src/bulk-upload-poc
RUN mkdir -p "$APP_HOME"

WORKDIR "$APP_HOME"

EXPOSE 8080
CMD ["make", "http"]
