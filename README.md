# Mail Service

[![Test Status](https://github.com/jo-hoe/go-mail-service/workflows/test/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/go-mail-service)](https://goreportcard.com/report/github.com/jo-hoe/go-mail-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/go-mail-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/go-mail-service?branch=main)

A simple mail service that allow to send mails.
Currently only [Sendgrid](https://sendgrid.com/) is implemented.

## Setup

### Pre-Requisites

- [Golang](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/engine/install/)

#### Optional

If you do not have it and run on Windows, you can directly install it from [gnuwin32](https://gnuwin32.sourceforge.net/packages/make.htm) or via `winget`

```PowerShell
winget install GnuWin32.Make
```

In case your want to deploy and access the service on k3d you will need to install the following tools:

- [K3d](https://k3d.io/v5.6.0/#releases)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [helm](https://helm.sh/docs/intro/install/)

Use `make` to run the project. Make is typically installed out of the box on Linux and Mac.

### Environment

Setup an .env file with the following content

```.env
API_PORT=80
IS_SENDGRID_ENABLED=true
IS_NOOP_ENABLED=false
DEFAULT_FROM_ADDRESS=<email address>
DEFAULT_FROM_NAME=<mail sender in clear text>
SENDGRID_API_KEY=<sendgrid api key>
```

## Run

### Docker

For plain docker run to the following commands

```bash
docker build . -t go-mail-service
docker run --rm -p 80:80 --env-file .env go-mail-service
```

### K3s

To run in k3d use the following command

```bash
make start-k3s
```

## Example Request

The service offers a basic API to send mails.
One can specify the subject, content, and addressed to send to.

```bash
curl -H "Content-Type: application/json" --data '{"subject":"my subject", "content":"my message", "to":["test@mail.de"]}' http://localhost:80/v1/sendmail
```

## Linting

The project used `golangci-lint` for linting.

### Installation

<https://golangci-lint.run/welcome/install/>

### Run Linting

Run the linting locally by executing.

```cli
golangci-lint run ./...
```
