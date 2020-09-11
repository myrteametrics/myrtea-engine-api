![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/myrteametrics/myrtea-engine-api)
[![Build Status](https://travis-ci.com/myrteametrics/myrtea-engine-api.svg)](https://travis-ci.com/myrteametrics/myrtea-engine-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/myrteametrics/myrtea-engine-api)](https://goreportcard.com/report/github.com/myrteametrics/myrtea-engine-api)
[![codecov](https://codecov.io/gh/myrteametrics/myrtea-engine-api/branch/master/graph/badge.svg)](https://codecov.io/gh/myrteametrics/myrtea-engine-api)
[![GitHub license](https://img.shields.io/github/license/myrteametrics/myrtea-engine-api)](https://github.com/myrteametrics/myrtea-engine-api/blob/master/LICENSE)


# Myrtea

Myrtea is a platform dedicated to the monitoring of business processes (in real-time or not), with a strong focus on decision assistance.

## Installation

### Pre-requisite

Myrtea needs two external backbone components to run :

* An instance of PostgreSQL v10+
* An instance or cluster of Elasticsearch v6.x (v7+ is currently not supported)

These two components can be easily started using docker containers :

```sh
# Quick single elasticsearch node startup
docker run -d --name myrtea-elasticsearch -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" docker.elastic.co/elasticsearch/elasticsearch:6.7.0

# Quick postgresql instance startup
docker run -d --name myrtea-postgres -p 5432:5432 -e POSTGRES_DB="postgres" -e POSTGRES_USER="postgres" -e POSTGRES_PASSWORD="postgres" -v "$PWD/resources/model.sql:/docker-entrypoint-initdb.d/1-init.sql" postgres:11.0-alpine
```

### From Source

Myrtea Engine-API requires Go version 1.14 or newer, the Makefile requires GNU make.

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
docker pull <myrtea-engine:v4.0.0>
docker run -d --name myrtea-engine -p 9000:9000 <myrtea-engine:v4.0.0>
```

This image can be configured using a configuration file mounted in the container with `-v $PWD/config/engine-api.toml:/app/config/engine-api.toml`.

It can also be configured using environment variables prefixed by `MYRTEA_`. (example: `-e MYRTEA_API_ENABLE_SECURITY=false` to disable security)

## Documentation

See the [Getting Started](https://myrteametrics.github.io/myrtea-docs/getting-started/first-application/) documentation for more infos on the application settings

## Why are we already in v4 !?

This product already has a pretty large history as a closed-source code base. The v4 is the first open-source release to date.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[![GitHub license](https://img.shields.io/github/license/myrteametrics/myrtea-engine-api)](https://github.com/myrteametrics/myrtea-engine-api/blob/master/LICENSE)
