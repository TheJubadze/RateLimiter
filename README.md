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