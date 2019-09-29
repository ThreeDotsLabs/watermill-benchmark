# Watermill Benchmark
<img align="right" width="200" src="https://threedots.tech/watermill-io/watermill-logo.png">

This is an early set of tools for benchmarking [watermill](https://github.com/ThreeDotsLabs/watermill).

**Warning:** This tool is meant to provide a rough estimate on how fast each Pub/Sub can process messages.
It uses very simplified infrastructure to set things up and default configurations.

Keep in mind that final performance depends on multiple factors.

## How it works

* All tests are run on a single 16 CPU GCloud compute instance (`n1-highcpu-16`).
* Docker Compose is used to run Pub/Sub infrastructure and benchmark code (except for Google Cloud Pub/Sub).
* The tool will first produce a big number of messages on a generated topic.
* Then it will subscribe to the topic and consume all of the messages.
* Multiple message sizes can be chosen (by default: 16, 64 and 256 bytes).

## Results

This is an early version of benchmark results. Expect it to be updated and grow.

### Kafka (one node)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 63506                  | 110811                   |
| 64                   | 57384                  | 110269                   |
| 256                  | 53744                  | 114685                   |

### Kafka (five nodes)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 70252                  | 117529                   |
| 64                   | 59801                  | 118052                   |
| 256                  | 55574                  | 121621                   |

### NATS

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 76208                  | 38169                    |
| 64                   | 57311                  | 39392                    |
| 256                  | 51209                  | 37670                    |

### SQL (MySQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 7299                   | 154                      |
| 64                   | 7354                   | 152                      |
| 256                  | 7253                   | 150                      |

### SQL (PostgreSQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 4142                   | 98                       |
| 64                   | 4040                   | 92                       |
| 256                  | 3933                   | 78                       |

### Google Cloud Pub/Sub (16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   |7416                    | 39591                    |
| 64                   |7555                    | 45653                    |
| 256                  |6967                    | 44483                    |

### AMQP (RabbitMQ, 16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 2408                   | 10608                    |
| 64                   | 2401                   | 12933                    |
| 256                  | 2388                   | 8144                     |

### GoChannel

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 272938                 | 101371                   |
| 64                   | 346184                 | 122439                   |
| 256                  | 333664                 | 130205                   |

## VM Setup

The project includes [Terraform](https://www.terraform.io/) definition for setting up an instance on Google Cloud Platform.

It will spin up a fresh Ubuntu 19.04 instance, install docker with dependencies and clone this repository.

Set environment variables:

```bash
# path to GCP credentials file
TF_VAR_credentials_path=
# project name on GCP
TF_VAR_project=
# public part of the key that you will use to access SSH
TF_VAR_pub_key_path=
```

Create the VM:

```bash
cd setup
terraform apply
```

The command will output the public IP address of the server. Use ssh with user `benchmark` to access it.

After running all benchmarks, destroy the VM:

```bash
terraform destroy
```

## Configuration

### Google Pub/Sub

Set environment variables in `compose/.env`:

```bash
# path to json file within the project with GCP credentials
GOOGLE_APPLICATION_CREDENTIALS=compose/key.json
# project name on GCP
GOOGLE_CLOUD_PROJECT=
```

## Running

Run benchmarks with:

```bash
./run.sh <pubsub>
```
