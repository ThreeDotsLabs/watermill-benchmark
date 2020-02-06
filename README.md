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
| 16                   | 63,506                 | 110,811                  |
| 64                   | 57,384                 | 110,269                  |
| 256                  | 53,744                 | 114,685                  |

### Kafka (five nodes)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 70,252                 | 117,529                  |
| 64                   | 59,801                 | 118,052                  |
| 256                  | 55,574                 | 121,621                  |

### NATS

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 76,208                 | 38,169                   |
| 64                   | 57,311                 | 39,392                   |
| 256                  | 51,209                 | 37,670                   |

### SQL (MySQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 7,299                  | 154                      |
| 64                   | 7,354                  | 152                      |
| 256                  | 7,253                  | 150                      |

### SQL (PostgreSQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 4,070                  | 396                      |
| 64                   | 3,895                  | 380                      |
| 256                  | 3,820                  | 378                      |

### Google Cloud Pub/Sub (16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 7,416                  | 39,591                   |
| 64                   | 7,555                  | 45,653                   |
| 256                  | 6,967                  | 44,483                   |

### AMQP (RabbitMQ, 16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 2,408                  | 10,608                   |
| 64                   | 2,401                  | 12,933                   |
| 256                  | 2,388                  | 8,144                    |

### GoChannel

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- | ---------------------- | ------------------------ |
| 16                   | 272,938                | 101,371                  |
| 64                   | 346,184                | 122,439                  |
| 256                  | 333,664                | 130,205                  |

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
