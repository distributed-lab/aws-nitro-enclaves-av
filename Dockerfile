FROM golang:1.23-alpine AS buildbase

RUN apk add git build-base ca-certificates

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN mkdir -p target/bin
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux go build -trimpath -buildvcs=false -ldflags="-s -w" -o target/bin/aws-nitro-enclaves-av .


FROM scratch
COPY --from=alpine:3.9 /bin/sh /bin/sh
COPY --from=alpine:3.9 /usr /usr
COPY --from=alpine:3.9 /lib /lib

COPY --from=buildbase /workspace/target/bin/aws-nitro-enclaves-av /usr/local/bin/aws-nitro-enclaves-av
COPY --from=buildbase /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["aws-nitro-enclaves-av"]
