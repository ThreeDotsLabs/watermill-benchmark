#!/bin/bash
set -e

readonly pubsub="$1"

if [ -z "$pubsub" ]; then
    echo "Usage: $0 <pubsub>"
    exit 1
fi

compose_flags=
if [ -f "./compose/$pubsub.yml" ]; then
    compose_flags="-f ./compose/$pubsub.yml"
    docker-compose $compose_flags up -d --remove-orphans

    # TODO replace with waiting for port
    sleep 20
fi

if [ ! -d ./vendor ]; then
    docker-compose -f ./compose/benchmark.yml run \
        -v "$(pwd):/benchmark" \
        -w /benchmark \
        benchmark go mod vendor
fi

docker-compose $compose_flags -f ./compose/benchmark.yml run \
    -v "$(pwd):/benchmark" \
    -w /benchmark \
    benchmark go run -mod=vendor ./cmd/main.go -pubsub "$pubsub"
