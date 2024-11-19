# Watermill Benchmark
<img align="right" width="300" src="https://watermill.io/img/gopher.svg">

This is a set of tools for benchmarking [watermill](https://github.com/ThreeDotsLabs/watermill).

**Warning:** This tool is meant to provide a rough estimate on how fast each Pub/Sub can process messages.
It uses very simplified infrastructure to set things up and default configurations.

Keep in mind that final performance depends on multiple factors.

**It's not meant to be a definitive answer on which Pub/Sub is the fastest.**
It should give you an idea of the ballpark performance you can expect.

## How it works

* All tests are run on a single 16 CPU GCloud compute instance (`n1-highcpu-16`).
* Docker Compose is used to run Pub/Sub infrastructure and benchmark code (except for Google Cloud Pub/Sub).
* The tool will first produce a big number of messages on a generated topic.
* Then it will subscribe to the topic and consume all of the messages.
* Multiple message sizes can be chosen (by default: 16, 64 and 256 bytes).

## Results (as of 18 November 2024)

### Kafka (one node)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 41,492                 | 101,669                  |
| 64                   | 40,189                 | 106,264                  |
| 256                  | 40,044                 | 107,278                  |

### NATS Jetstream (16 Subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) | Subscribe (messages / s - async ack) |
|----------------------|------------------------|--------------------------|--------------------------------------|
| 16                   | 50,668                 | 34,713                   | 59,728                               |
| 64                   | 49,204                 | 34,561                   | 59,743                               |
| 256                  | 48,242                 | 34,097                   | 59,385                               |

### NATS Jetstream (48 Subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages /s ) | Subscribe (messages / s - async ack) |
|----------------------|------------------------|--------------------------|--------------------------------------|
| 16                   | 50,680                 | 46,377                   | 86,348                               |
| 64                   | 49,341                 | 46,307                   | 86,078                               |
| 256                  | 48,744                 | 46,035                   | 86,499                               |

### Redis

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 59,158                 | 12,134                   |
| 64                   | 58,988                 | 12,392                   |
| 256                  | 58,038                 | 12,133                   |

### SQL (MySQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s - batch size = 1) | Subscribe (messages / s - batch size = 100) |
|----------------------|------------------------|-------------------------------------------|---------------------------------------------|
| 16                   | 6,371                  | 283                                       | 2,794                                       |
| 64                   | 9,887                  | 281                                       | 2,637                                       |
| 256                  | 9,596                  | 271                                       | 2,766                                       |

### SQL (PostgreSQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s - batch size = 1) | Subscribe (messages / s - batch size = 100) | 
|----------------------|------------------------|-------------------------------------------|---------------------------------------------|
| 16                   | 2,552                  | 122                                       | 9,460                                       |
| 64                   | 2,831                  | 118                                       | 9,045                                       |
| 256                  | 2,744                  | 104                                       | 7,843                                       |

### SQL (PostgreSQL Queue)

| Subscribe Batch Size | Message size (bytes) | Publish (messages / s) |  Subscribe (messages / s - batch size = 1) | Subscribe (messages / s - batch size = 100) | 
|----------------------|----------------------|------------------------|--------------------------------------------|---------------------------------------------|
| 100                  | 16                   | 2,825                  | 146                                        | 10,466                                      |
| 100                  | 64                   | 2,842                  | 147                                        | 9,626                                       |
| 100                  | 256                  | 2,845                  | 138                                        | 8,276                                       |

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 2,825                  | 0,146                    |
| 64                   | 2,842                  | 0,147                    |
| 256                  | 2,845                  | 0,138                    |

### Google Cloud Pub/Sub (16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 3,027                  | 28,589                   |
| 64                   | 3,020                  | 31,057                   |
| 256                  | 2,918                  | 32,722                   |

### AMQP (RabbitMQ, 16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 2,770                  | 14,604                   |
| 64                   | 2,752                  | 12,128                   |
| 256                  | 2,750                  | 8,550                    |

### GoChannel

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 315,776                | 138,743                  |
| 64                   | 325,341                | 163,034                  |
| 256                  | 341,223                | 145,718                  |

## VM Setup

The project includes [Terraform](https://www.terraform.io/) definition for setting up an instance on Google Cloud Platform.

It will spin up a fresh Ubuntu 19.04 instance, install docker with dependencies and clone this repository.

Set environment variables:

```bash
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
