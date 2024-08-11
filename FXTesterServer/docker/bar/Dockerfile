FROM golang:1.22.1-bullseye as builder

WORKDIR /go/src
COPY go.* ./
RUN go mod download
COPY . .
RUN go install ./cmd/bar/

FROM ubuntu:22.04
RUN apt update && apt install -y ca-certificates && apt clean
COPY --from=builder /go/bin/bar /
COPY ./settings/config.yaml /settings/
COPY ./settings/cert/fx-tester.cert.pem /settings/cert/
COPY ./settings/cert/fx-tester.key.pem /settings/cert/
COPY ${ROOT_CA_PATH} /usr/local/share/ca-certificates/ca.cert.crt
RUN update-ca-certificates
ENTRYPOINT ["/bar"]