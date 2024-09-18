# Rate Limiter

[![Build Status](https://github.com/TheJubadze/RateLimiter/actions/workflows/tests.yml/badge.svg)](https://github.com/TheJubadze/RateLimiter/actions/workflows/tests.yml)

## Overview

Rate Limiter is a Go-based application designed to manage and enforce rate limiting policies for various resources.

## Features

- IP Whitelisting and Blacklisting
- Rate limiting based on IP, login, and password
- gRPC API for integration

## Getting Started

### Prerequisites

- Go 1.22+
- Docker
- Docker Compose

### Building the Project

To build the project, run:

```sh
make build
```

### Running the Project in Docker

To run the project, run:

```sh
make up
```

### Running the Tests

To run the tests, run:

```sh
make test
```

### Running the Linter

To run the linter, run:

```sh
make lint
```

### Running the Integration Tests

To run the integration tests, run:

```sh
make integration-tests
```
