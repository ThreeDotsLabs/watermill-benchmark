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
| -------------------- |------------------------|--------------------------|
| 16                   | 44,090                 | 108.285                  |
| 64                   | 41,371                 | 108,848                  |
| 256                  | 41,497                 | 111,756                  |

### NATS Jetstream (16 Subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) | Subscribe (messages / s - async ack) |
|----------------------|------------------------|--------------------------|--------------------------------------|
| 16                   | 49,255                 | 33,009                   | 63,065                               |
| 64                   | 49,296                 | 33,016                   | 62,667                               |
| 256                  | 48.488                 | 32,573                   | 62,745                               |

### NATS Jetstream (48 Subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) | Subscribe (messages / s - async ack) |
|----------------------|------------------------|--------------------------|--------------------------------------|
| 16                   | 48,882                 | 45,275                   | 90,917                               |
| 64                   | 48,681                 | 44,746                   | 89,527                               |
| 256                  | 48,097                 | 44,487                   | 90,510                               |

### Redis

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 61,642                 | 11,213                   |
| 64                   | 58,554                 | 11,252                   |
| 256                  | 58,906                 | 11,408                   |

### SQL (MySQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 5,599                  | 167                      |
| 64                   | 5,625                  | 168                      |
| 256                  | 5,381                  | 164                      |

### SQL (PostgreSQL)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 3,834                  | 455                      |
| 64                   | 3,923                  | 468                      |
| 256                  | 3,855                  | 472                      |

### Google Cloud Pub/Sub (16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
| -------------------- |------------------------|--------------------------|
| 16                   | 3,689                  | 30,229                   |
| 64                   | 3,408                  | 26,448                   |
| 256                  | 6,967                  | 30,123                   |

### AMQP (RabbitMQ, 16 subscribers)

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 2,702                  | 13,192                   |
| 64                   | 2,712                  | 12,980                   |
| 256                  | 2,692                  | 8,027                    |

### GoChannel

| Message size (bytes) | Publish (messages / s) | Subscribe (messages / s) |
|----------------------|------------------------|--------------------------|
| 16                   | 331,882                | 118,943                  |
| 64                   | 298,847                | 123,499                  |
| 256                  | 373,053                | 130,940                  |

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
