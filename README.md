![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/myrteametrics/myrtea-engine-api)
[![Build Status](https://github.com/myrteametrics/myrtea-engine-api/actions/workflows/go.yml/badge.svg)](https://github.com/myrteametrics/myrtea-engine-api/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/myrteametrics/myrtea-engine-api)](https://goreportcard.com/report/github.com/myrteametrics/myrtea-engine-api)
[![codecov](https://codecov.io/gh/myrteametrics/myrtea-engine-api/branch/master/graph/badge.svg)](https://codecov.io/gh/myrteametrics/myrtea-engine-api)
[![GitHub license](https://img.shields.io/github/license/myrteametrics/myrtea-engine-api)](https://github.com/myrteametrics/myrtea-engine-api/blob/master/LICENSE)


# Myrtea

Myrtea is a platform dedicated to the monitoring of business processes (in real-time or not), with a strong focus on decision assistance.

## Installation

### Pre-requisite

Myrtea needs two external backbone components to run :

* An instance of PostgreSQL v10+
* An instance or cluster of Elasticsearch >= v7.0

These two components can be easily started using docker containers :

```sh
# Quick single elasticsearch node startup
docker run -d --name myrtea-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:7.17.28

# Quick postgresql instance startup
docker run -d --name myrtea-postgres -p 5432:5432 -e POSTGRES_DB="postgres" -e POSTGRES_USER="postgres" -e POSTGRES_PASSWORD="postgres"  postgres:16-alpine
```

### From Source

Myrtea Engine-API requires Go version 1.23 or newer, the Makefile requires GNU make.

```sh
git clone https://github.com/myrteametrics/myrtea-engine-api.git
cd ./myrtea-engine-api
make swag build
```

From there, you can either run the compiled binary with :

```sh
make run
```

Or generate a fresh and ready to use docker image :

```sh
make docker-build-local
```

### With Docker

Our favored way to ship and install Myrtea is by using [docker](https://www.docker.com/) images and containers.

```sh
# This docker image might not be available as you read this (and you might need to build it yourself following the "From Source" section)
docker pull <myrtea-engine:v5.0.0>
docker run -d --name myrtea-engine -p 9000:9000 <myrtea-engine:v5.0.0>
```

This image can be configured using a configuration file mounted in the container with `-v $PWD/config/engine-api.toml:/app/config/engine-api.toml`.

It can also be configured using environment variables prefixed by `MYRTEA_`. (example: `-e MYRTEA_API_ENABLE_SECURITY=false` to disable security)

### With Docker Compose

You can also build a myrtea server instance in a single command with docker compose. 
Two compose files are included in the project: dev and prod (default).

As we are in docker containers, please replace all `localhost` occurrences 
in the `config/engine-api.toml` file with the container name linking to it.

Please make sure you have done all the required configurations.

#### Production

You can create an .env file containing the following environment variables:
```
POSTGRES_DB: db
POSTGRES_USER: pg-user
POSTGRES_PASSWORD: pg-pass
```

Then enter the command:

```sh
docker-compose --env-file .env up
```

#### Development

No need to configure anything else here, just enter the following command:

```sh
docker-compose -f docker-compose.dev.yml up
```

## Documentation

See the [Getting Started](https://myrteametrics.github.io/myrtea-docs/getting-started/first-application/) documentation for more infos on the application settings

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[![GitHub license](https://img.shields.io/github/license/myrteametrics/myrtea-engine-api)](https://github.com/myrteametrics/myrtea-engine-api/blob/master/LICENSE)
