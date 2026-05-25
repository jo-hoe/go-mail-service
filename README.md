# Mail Service

[![Test Status](https://github.com/jo-hoe/go-mail-service/workflows/test/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/go-mail-service)](https://goreportcard.com/report/github.com/jo-hoe/go-mail-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/go-mail-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/go-mail-service?branch=main)

A mail service exposing both a REST API and an SMTP listener, backed by pluggable providers.

**Providers:**

- [SendGrid](https://sendgrid.com/)
- [Mailjet](https://www.mailjet.com/)
- Noop (logs mail without sending — development only)

**Interfaces:**

- HTTP `POST /v1/sendmail` (proprietary REST API, default port `8080`)
- SMTP listener with optional `AUTH PLAIN` and STARTTLS (default port `587`)

## Go Client Library

A Go client library is available for easy integration with your Go applications:

```bash
go get github.com/jo-hoe/go-mail-service/pkg/client
```

### Quick Start with Client

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/jo-hoe/go-mail-service/pkg/client"
)

func main() {
    // Create a new client
    mailClient := client.NewClient("http://localhost:8080")

    // Create a mail request
    request := client.MailRequest{
        To:          "recipient@example.com",
        Subject:     "Hello from Go Mail Service",
        HtmlContent: "<h1>Hello!</h1><p>This is a test email.</p>",
        From:        "sender@example.com",
        FromName:    "Test Sender",
    }

    // Send the email
    response, err := mailClient.SendMail(context.Background(), request)
    if err != nil {
        log.Fatalf("Failed to send email: %v", err)
    }

    fmt.Printf("Email sent successfully to: %s\n", response.To)
}
```

For detailed documentation, see [pkg/client/README.md](pkg/client/README.md).

## Setup

### Git Hooks

This project uses git hooks to ensure code quality. After cloning the repository, run the setup script to configure the hooks:

```bash
# On Linux/Mac
sh scripts/setup-hooks.sh

# On Windows (Git Bash)
sh scripts/setup-hooks.sh

# Or manually configure
git config core.hooksPath scripts/git-hooks
```

**Active Hooks:**

- **pre-commit**: Automatically runs `go fmt ./...` before each commit to ensure consistent code formatting

### Prerequisites

- [Golang](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/engine/install/)

#### Optional

If you do not have it and run on Windows, you can directly install it from [gnuwin32](https://gnuwin32.sourceforge.net/packages/make.htm) or via `winget`

```PowerShell
winget install GnuWin32.Make
```

In case you want to deploy and access the service on k3d you will need to install the following tools:

- [K3d](https://k3d.io/v5.6.0/#releases)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [helm](https://helm.sh/docs/intro/install/)

Run the project using `make`. Make is typically installed by default on Linux and Mac.

## Configuration

The service reads a single YAML file at the path given by `CONFIG_PATH` (default `/config/config.yaml`). Secrets are read from files mounted at paths referenced in the config — **the service does not consume any other environment variables**.

### Config file shape

```yaml
logLevel: "info"          # debug | info | warn | error

sender:
  address: "noreply@example.com"
  name: "My Service"

http:
  port: 8080

smtp:
  port: 587
  domain: "mail.example.com"     # advertised in EHLO
  auth:
    required: true
    username: "smtp-user"
    passwordFile: "/secrets/smtp/password"
  tls:
    enabled: false
    certFile: ""
    keyFile: ""

provider:
  # Enable exactly one. Priority if multiple are enabled: mailjet > sendgrid > noop.
  mailjet:
    enabled: false
    apiKeyPublicFile:  "/secrets/mailjet/apiKeyPublic"
    apiKeyPrivateFile: "/secrets/mailjet/apiKeyPrivate"
  sendgrid:
    enabled: false
    apiKeyFile: "/secrets/sendgrid/apiKey"
  noop:
    enabled: false
```

A ready-to-run example with the noop provider lives at `local/config.yaml`.

> ⚠️ The **noop** provider logs full mail details (recipients, subject, body) and is for development/testing only. Never enable it in production.

### Local Makefile workflow

The Makefile uses a `.env` file to feed `helm --set` flags during local k3d deployment. The Go app itself does not read these variables.

```bash
cp .env.example .env   # then edit values
```

## Run

### Plain Docker (uses local/config.yaml)

```bash
make start-docker
```

This builds the image, mounts `local/config.yaml` to `/config/config.yaml`, and exposes both HTTP (`8080`) and SMTP (`587`).

### k3d

```bash
make start-k3d
```

Spins up a local k3d cluster, pushes the image to its registry, and deploys the Helm chart using values driven from `.env`.

## Example Requests

### HTTP REST API

```bash
curl -H "Content-Type: application/json" \
     --data '{"subject":"my subject","content":"my message","to":"test@mail.com,test2@mail.com"}' \
     http://localhost:8080/v1/sendmail
```

### SMTP

```bash
# Without auth (smtp.auth.required: false)
swaks --to recipient@example.com --from sender@example.com \
      --server localhost --port 587

# With AUTH PLAIN
swaks --to recipient@example.com --server localhost --port 587 \
      --auth-user smtp-user --auth-password secret

# Probe STARTTLS
openssl s_client -starttls smtp -connect localhost:587
```

## Linting

The project used `golangci-lint` for linting.

### Installation

<https://golangci-lint.run/welcome/install/>

### Run Linting

Run the linting locally by executing.

```bash
make lint
```

The lint configuration lives in `.golangci.yml` (linters: `dupl`, `gocyclo`, `gosec`, `misspell`).
