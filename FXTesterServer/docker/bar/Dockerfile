FROM golang:1.22.1-bullseye as builder

WORKDIR /go/src
COPY go.* ./
RUN go mod download
COPY . .
RUN go install ./cmd/bar/

FROM ubuntu:22.04
RUN apt update && apt install -y ca-certificates && apt clean
COPY --from=builder /go/bin/bar /
COPY ./cert/myservice.cert /go/bin/
COPY ./cert/myservice.key /go/bin/
COPY ./cert/fx-tester.key.pem /go/bin/key.pem
COPY ./cert/fx-tester.cert.pem /go/bin/cert.pem
COPY ${ROOT_CA_PATH} /usr/local/share/ca-certificates/ca.cert.crt
RUN update-ca-certificates
ENTRYPOINT ["/bar"]