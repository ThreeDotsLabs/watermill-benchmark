#!/bin/bash
set -e

readonly compose="$1"

if [ -z "$compose" ]; then
    echo "Usage: $0 <compose_name>"
    exit 1
fi

compose_flags=
if [ -f "./compose/$compose.yml" ]; then
    compose_flags="-f ./compose/$compose.yml"
    docker compose $compose_flags up -d --remove-orphans

    # TODO replace with waiting for port
    sleep 20
fi

if [ ! -d ./vendor ]; then
    docker compose -f ./compose/benchmark.yml run \
        -v "$(pwd):/benchmark" \
        -w /benchmark \
        benchmark go mod vendor
fi

docker compose $compose_flags -f ./compose/benchmark.yml run \
    -v "$(pwd):/benchmark" \
    -w /benchmark \
    benchmark go run -mod=vendor ./cmd/main.go -pubsub "$compose"
