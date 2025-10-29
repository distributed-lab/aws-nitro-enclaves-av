# aws-nitro-enclaves-av

## Description

[//]: # (TODO: add description)

## Build
  Make sure that `docker` and `nitro-cli` installed and that processor architecture where you are building the service is x86-64 (amd64).

  See [How to run](#how-to-run) to install docker.

  Clone repo:
  ```bash
  git clone https://github.com/distributed-lab/aws-nitro-enclaves-av.git
  cd aws-nitro-enclaves-av
  ```

  Build Enclave Image File:
  ```bash
  ./build.sh
  ```

  The file can be found in ./output

  Build Dockerfile:
  ```bash
  ./build.docker.sh
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

3. Install `nitro-cli` and `socat`:
```bash
yum install aws-nitro-enclaves-cli aws-nitro-enclaves-cli-devel socat -y
usermod -aG ne ec2-user
usermod -aG docker ec2-user
```

4. Configure necessary amount of CPU and RAM for service in `/etc/nitro_enclaves/allocator.yaml`

5. Start `docker` and `allocator` services:
```bash
systemctl enable --now nitro-enclaves-allocator.service
systemctl enable --now docker
```

### Running

1. Enable kernel modules
```bash
sudo modprobe nfs
sudo modprobe nfsd
```

2. Run docker compose (it takes ~ 6 minutes)
```bash
docker compose up
```

3. If this is not the first launch, the attestation documents from the previous launch must be placed in `volume/attestations`. If any documents are missing, they will be automatically generated in the following sequence: `kms_key_id.coses1` -> `private_key.coses1` -> `public_key.coses1` -> `address.coses1`.

If this is the first launch, you can find the generated attestation documents in the `volume/attestations` directory.

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

Endpoints: `v1/attestation-documents/address` and `v1/attestation-documents/public-key` returns `address` and `public_key` of Attestation Verifier

## Testing
To run the tests, you need to repeat all the steps described in the [How to run](#how-to-run) section, except for actually launching the enclave.

You need to install golang on the EC2 instance.

Start service in enclave debug mode: Add extra option in docker-compose.yaml

Run tests:
```bash
go test ./tests
```
