# Results

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
| 16                   | 6989                   | 143                      |
| 64                   | 7322                   | 143                      |
| 256                  | 6897                   | 142                      |

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
