FROM debian:bookworm-slim@sha256:9852c9b122fa2dce95ea33a096292ce649a12a7ff321a6a6f1a40eca4989a9fc AS socat-builder
COPY sources.list /etc/apt/sources.list
RUN export DEBIAN_FRONTEND=noninteractive && \
    apt-get update && \
    apt-get install -y --no-install-recommends \
    wget make gcc build-essential && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
RUN wget http://www.dest-unreach.org/socat/download/socat-1.7.4.4.tar.gz
RUN echo "0f8f4b9d5c60b8c53d17b60d79ababc4a0f51b3bb6d2bd3ae8a6a4b9d68f195e socat-1.7.4.4.tar.gz" | sha256sum -c -
RUN tar -xzf socat-1.7.4.4.tar.gz && \
    cd socat-1.7.4.4 && \
    ./configure && \
    make && \
    make install


FROM golang:1.23-alpine@sha256:cf956b65299c780e1035dcb13aa755a9ad08a3aa09b8cb54d66e61c8273fb828 AS buildbase

RUN apk add git build-base

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN mkdir -p target/bin
RUN CGO_ENABLED=0 GO111MODULE=on GOOS=linux go build -trimpath -buildvcs=false -ldflags="-s -w" -o target/bin/aws-nitro-enclaves-av .


FROM debian:bookworm-slim@sha256:9852c9b122fa2dce95ea33a096292ce649a12a7ff321a6a6f1a40eca4989a9fc

COPY sources.list /etc/apt/sources.list
RUN export DEBIAN_FRONTEND=noninteractive && \
    apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates iproute2 nfs-common && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

COPY --from=buildbase /workspace/target/bin/aws-nitro-enclaves-av /usr/local/bin/aws-nitro-enclaves-av
COPY --from=socat-builder /usr/local/bin/socat /usr/local/bin/

COPY run.sh /root/run.sh
RUN chmod +x /root/run.sh

ENTRYPOINT [ "/root/run.sh" ]
