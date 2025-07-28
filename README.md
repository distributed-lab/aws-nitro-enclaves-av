# aws-nitro-enclaves-av

## Description

[//]: # (TODO: add description)

## Install

  ```
  git clone github.com/distributed-lab/aws-nitro-enclaves-av
  cd aws-nitro-enclaves-av
  go build main.go
  export KV_VIPER_FILE=./config.yaml
  ./main run service
  ```

## Documentation

We do use openapi:json standard for API. We use swagger for documenting our API.

To open online documentation, go to [swagger editor](http://localhost:8080/swagger-editor/) here is how you can start it
```
  cd docs
  npm install
  npm start
```
To build documentation use `npm run build` command,
that will create open-api documentation in `web_deploy` folder.

To generate resources for Go models run `./generate.sh` script in root folder.
use `./generate.sh --help` to see all available options.

Note: if you are using Gitlab for building project `docs/spec/paths` folder must not be
empty, otherwise only `Build and Publish` job will be passed.  

## Running from docker 
  
Make sure that docker installed.

use `docker run ` with `-p 8080:80` to expose port 80 to 8080

  ```
  docker build -t github.com/distributed-lab/aws-nitro-enclaves-av .
  docker run -e KV_VIPER_FILE=/config.yaml github.com/distributed-lab/aws-nitro-enclaves-av
  ```

## Running from Source

* Set up environment value with config file path `KV_VIPER_FILE=./config.yaml`
* Provide valid config file
* Launch the service with `run service` command


## Documentation
Endpoint: `v1/attestations`
### Request
```json
{
  "type": "attestations",
  "attributes": {
    "attestation": "string",
    "domain": {},
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
