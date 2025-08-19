# aws-nitro-enclaves-av

## Description

[//]: # (TODO: add description)

## Build
  Make sure that `docker` and `nitro-cli` installed and that processor architecture where you are building the service is x86-64 (amd64).

  Clone repo:
  ```
  git clone https://github.com/distributed-lab/aws-nitro-enclaves-av.git
  cd aws-nitro-enclaves-av
  ```

  Build container:

  ```
  docker build -t github.com/distributed-lab/aws-nitro-enclaves-av .
  ```

  Build Enclave Image File:
  ```
  nitro-cli build-enclave --docker-uri github.com/distributed-lab/aws-nitro-enclaves-av:latest --output-file attestation-verifier.eif
  ```

## How to run

### Preparation

1. Create an IAM role with the following policies:
```
kms:Decrypt
kms:CreateKey
sts:GetCallerIdentity
kms:GenerateDataKeyPair
```

2. Create EC2 instance with `Amazon Linux 2023 x86-64` and `Nitro Enclaves: Enabled`

3. Install `nitro-cli`:
```bash
yum install aws-nitro-enclaves-cli -y
yum install aws-nitro-enclaves-cli-devel -y
usermod -aG ne ec2-user
usermod -aG docker ec2-user
```

4. Configure necessary amount of CPU and RAM for service in `/etc/nitro_enclaves/allocator.yaml`

5. Start `docker` and `allocator` services:
```bash
systemctl enable --now nitro-enclaves-allocator.service
systemctl enable --now docker
```

6. Install `socat`:
```bash
yum install -y wget tar gcc
wget http://www.dest-unreach.org/socat/download/socat-1.7.4.4.tar.gz
tar -xzf socat-1.7.4.4.tar.gz
cd socat-1.7.4.4
./configure
make
make install
```

7. Setup directory for storing service persistent files:
```bash
export SERVICE_DIR="/export/attestation-verifier"
chmod 755 $SERVICE_DIR
chown -R ec2-user:ec2-user $SERVICE_DIR
```

8. Install, configure and start `nfs`:
```bash
yum install -y nfs-utils
echo "$SERVICE_DIR 127.0.0.1/32(rw,insecure,fsid=0,crossmnt,no_subtree_check,sync,no_root_squash)" >> /etc/exports
systemctl restart nfs-server
systemctl enable nfs-server
```

### Running

1. Copy `config.yaml` in `$SERVICE_DIR`

2. If this is not the first launch, the attestation documents from the previous launch must be placed in `$SERVICE_DIR/attestations`. If any documents are missing, they will be automatically generated in the following sequence: `kms_key_id.coses1` -> `private_key.coses1` -> `public_key.coses1` -> `address.coses1`.

3. Start `socat` vsock proxies:
```bash
#!/bin/bash
TOKEN=`curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600"`
REGION=`curl -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/placement/region`

# AWS KMS
socat VSOCK-LISTEN:8002,fork,keepalive TCP:kms.$REGION.amazonaws.com:443,keepalive &

# AWS STS
socat VSOCK-LISTEN:8003,fork,keepalive TCP:sts.$REGION.amazonaws.com:443,keepalive &

# NFS Server
socat VSOCK-LISTEN:20000,fork,keepalive TCP:127.0.0.1:2049,keepalive &

# IMDS
socat VSOCK-LISTEN:16900,fork,keepalive TCP:169.254.169.254:80,keepalive &

# Service
readonly SERVICE_CID=16
readonly EC2_PORT=8000
socat TCP-LISTEN:$EC2_PORT,fork,reuseaddr,keepalive,bind=127.0.0.1 VSOCK-CONNECT:$SERVICE_CID:8080,keepalive &
```

You can change the context ID of the service to the one specified when launching the enclave. `EC2_PORT` is the port of the EC2 instance that will be redirected to the enclave.

4. Start enclave:
```bash
nitro-cli run-enclave --cpu-count 2 --memory 1024 --enclave-cid 16 --eif-path attestation-verifier.eif
```

If this is the first launch, you can find the generated attestation documents in the `$SERVICE_DIR/attestations` directory.

## Documentation
Endpoint: `v1/attestations`
### Request
```json
{
  "type": "attestations",
  "attributes": {
    "attestation": "string",
    "domain": {
      "name": "Test",
      "version": "1"
    },
    "primary_type": "Mail",
    "fields_to_sign": [
      "pcr0",
      "public_key"
    ]
  }
}
```

- `attestation` is standard base64-encoded AWS Nitro Enclave attestation document;
- `domain` is EIP712 domain like:
  ```json
  {
    "name": "My amazing dApp",
    "version": "2",
    "chainId": "1",
    "verifyingContract": "0x1c56346cd2a2bf3202f771f50d3d14a367b48070",
    "salt": "0x43efba6b4ccb1b6faa2625fe562bdd9a23260359"
  }
  ```
  All field is optional as specified in [EIP712](https://eips.ethereum.org/EIPS/eip-712), but `domain` field is required;
- `primary_type` is name of abstract structur. For example, `Mail(address to)` where `Mail` is primary type. Optional with default value `Register`;
- `fields_to_sign` - `pcrX` it is wildcard for `pcr0`, `pcr1`, ..., `pcr31`. Fields to sign is fields that will be included in EIP712 signature. For example: `Register(bytes pcr0,bytes public_key)` for `pcr0` and `public_key` fields. `pcrX`, `public_key`, `user_data` and `nonce` - bytes; `module_id` and `digest` - string; `timestamp` - uint64; Optional with default value `[ "pcr0", "public_key" ]`

### Response
```json
{
  "data": {
    "type": "attestations",
    "attributes": {
      "signature": "string"
    }
  }
}
```
`signature` is standard base64-encoded EIP712 signature.

## Testing
To run the tests, you need to repeat all the steps described in the [How to run](#how-to-run) section, except for actually launching the enclave.

You need to install golang on the EC2 instance and copy the attestation documents from `tests/attestations` to `$SERVICE_DIR/attestations`

Start service in enclave debug mode:
```bash
nitro-cli run-enclave --cpu-count 2 --memory 1024 --enclave-cid 16 --eif-path attestation-verifier.eif --debug-mode --attach-console
```

Run tests:
```bash
go test ./tests
```
