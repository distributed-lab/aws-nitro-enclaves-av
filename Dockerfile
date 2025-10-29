FROM golang:1.23-bookworm@sha256:e87b2a5f6df2dff71ea330d55d54f4979eb380ae58a7e3aabc9d53121243e689 AS buildbase

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN mkdir -p target/bin
RUN CGO_ENABLED=1 GO111MODULE=on GOOS=linux go build -trimpath -buildvcs=false -ldflags="-s -w" -o target/bin/attestation-verifier .


FROM debian:bookworm-slim@sha256:9852c9b122fa2dce95ea33a096292ce649a12a7ff321a6a6f1a40eca4989a9fc AS attestation-verifier-base

COPY enclave/sources.list /etc/apt/sources.list
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update
RUN apt-get install -y --no-install-recommends \
    ca-certificates \
    iproute2 \
    nfs-common \
    socat \
    curl && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=buildbase /workspace/target/bin/attestation-verifier /usr/local/bin/attestation-verifier

COPY enclave/run.sh /opt/run.sh
RUN chmod +x /opt/run.sh

ENTRYPOINT [ "/opt/run.sh" ]

FROM public.ecr.aws/amazonlinux/amazonlinux:2023
RUN yum install aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel socat -y

COPY output/attestation-verifier.eif /root
COPY run.sh /opt
COPY prestop-hook.sh /opt
RUN chmod +x /opt/run.sh
RUN chmod +x /opt/prestop-hook.sh

ENTRYPOINT [ "/opt/run.sh" ]