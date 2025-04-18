# IoTeX W3bstream Sprout Node Operator Guide

W3bstream is a permissionless, decentralized protocol within the IoTeX Network, where node operators contribute computing power to support verifiable computations for blockchain applications. These applications rely on insights from real-world data to trigger their token economies. Anyone can become a W3bstream Node Operator in the IoTeX Network, choosing which dApps to support in processing data and generating ZK (Zero Knowledge) Proofs.

## Run a W3bstream Node

The recommended method to run a W3bstream node is using official Docker images from IoTeX.

### Prerequisites

- Docker Engine (version 18.02 or higher):

  Check your Docker version:

  ```bash
  docker version
  ```

  [Docker installation instructions →](https://docs.docker.com/engine/install/)

- Docker Compose Plugin
  
  Verify Docker Compose installation:

  ```bash
  docker compose version # Install with → sudo apt install docker-compose-plugin
  ```

- **Blockchain Wallet**: A funded wallet on the target blockchain is required for your W3bstream node to dispatch proofs to blockchain contracts. For IoTeX Testnet, see [create a wallet](https://docs.iotex.io/the-iotex-stack/wallets/metamask), and claim test IOTX via [faucet](https://docs.iotex.io/the-iotex-stack/iotx-faucets/testnet-tokens#the-iotex-developer-portal).

- **Bonsai API Key**: If you are joining a project requiring RISC0 snark proofs, as the W3bstream protocol currently utilizes the [Bonsai API](https://dev.risczero.com/api/bonsai/), obtain an [API key here](https://docs.google.com/forms/d/e/1FAIpQLSf9mu18V65862GS4PLYd7tFTEKrl90J5GTyzw_d14ASxrruFQ/viewform).

### Download Docker Images

Fetch the latest stable docker-compose.yaml:

```bash
curl https://raw.githubusercontent.com/machinefi/sprout/release/docker-compose.yaml > docker-compose.yaml
```

Pull the required images:

```bash
docker compose pull
```

### Configure your blockchain account

To enable your node to send proofs to the destination blockchain, set up a funded account on the target chain:

```bash
export PRIVATE_KEY=${your private key}
```

### Optional: Provide your BONSAI API Key

For projects using RISC0 Provers, supply your Bonsai API Key:

```bash
export BONSAI_KEY=${your bonsai key}
```

Refer to the W3bstream project documentation for the dApp you are joining to determine if Risc Zero proofs are required.

### Manage the project file

You can move your project file to a directory, and mount it to W3bstream node container. The environment variable used to indicate a directory is **PROJECT_FILE_DIRECTORY**  
The default project directory is **./test/project**

### Manage the node

To start W3bstream, run the following command in the directory containing `docker-compose.yaml`:

```bash
docker compose up -d
```

Monitor the W3bstream instance status:

```bash
docker-compose logs -f coordinator prover
```

To shut down the W3bstream instance:

```bash
docker-compose down
```

#### Sending messages to the node

Send a message to a RISC0-based test project (ID 10000):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"projectID": 10000,"projectVersion": "0.1","data": "{\"private_input\":\"14\", \"public_input\":\"3,34\", \"receipt_type\":\"Snark\"}"}' http://localhost:9000/message
```

Send a message to the Halo2-based test project (ID 10001):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"projectID": 10001,"projectVersion": "0.1","data": "{\"private_a\": 3, \"private_b\": 4}"}' http://localhost:9000/message
```

Send a message to a zkWasm-based test project (ID 10002):

```bash
curl -X POST -H "Content-Type: application/json" -d '{"projectID": 10002,"projectVersion": "0.1","data": "{\"private_input\": [1, 1] , \"public_input\": [2] }"}' http://localhost:9000/message
```

#### Query the status of a proof request

After sending a message, you'll receive a message ID as a response from the node, e.g.,

```json
{
 "messageID": "8785a42c-9d6c-4780-910c-de0147aea243"
}
```

you can quesry the history of the proof request with:

```bash
curl http://localhost:9000/message/8785a42c-9d6c-4780-910c-de0147aea243 | jq -r '.'
```

example result:

```json
{
  "messageID": "8785a42c-9d6c-4780-910c-de0147aea243",
  "states": [
    {
      "state": "received",
      "time": "2024-06-10T09:30:05.790151Z",
      "comment": "",
      "result": ""
    },
    {
      "state": "packed",
      "time": "2024-06-10T09:30:05.793218Z",
      "comment": "",
      "result": ""
    },
    {
      "state": "dispatched",
      "time": "2024-06-10T09:30:10.87987Z",
      "comment": "",
      "result": ""
    },
    {
      "state": "proved",
      "time": "2024-06-10T09:30:11.193027Z",
      "comment": "",
      "result": "proof result"
    },
    {
      "state": "outputted",
      "time": "2024-06-10T09:30:11.20942Z",
      "comment": "output type: stdout",
      "result": ""
    }
  ]
}
```

When the request is in "proved" state, you can check out the node logs to find out the hash of the blockchain transaction that wrote the proof to the destination chain.
