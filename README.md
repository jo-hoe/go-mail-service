# Mail Service

[![Test Status](https://github.com/jo-hoe/go-mail-service/workflows/test/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/go-mail-service)](https://goreportcard.com/report/github.com/jo-hoe/go-mail-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/go-mail-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/go-mail-service?branch=main)

A simple mail service that allows you to send mails.
Currently supports:

- [SendGrid](https://sendgrid.com/)
- [Mailjet](https://www.mailjet.com/)

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
    mailClient := client.NewClient("http://localhost:80")

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

### Environment

Create a `.env` file by copying the example file:

```bash
cp .env.example .env
```

Then edit the `.env` file with your actual configuration values.

#### SendGrid Configuration

```.env
API_PORT=80
IS_SENDGRID_ENABLED=true
IS_MAILJET_ENABLED=false
IS_NOOP_ENABLED=false
DEFAULT_FROM_ADDRESS=<email address>
DEFAULT_FROM_NAME=<mail sender in clear text>
SENDGRID_API_KEY=<sendgrid api key>
```

#### Mailjet Configuration

```.env
API_PORT=80
IS_SENDGRID_ENABLED=false
IS_MAILJET_ENABLED=true
IS_NOOP_ENABLED=false
DEFAULT_FROM_ADDRESS=<email address>
DEFAULT_FROM_NAME=<mail sender in clear text>
MAILJET_API_KEY_PUBLIC=<mailjet public api key>
MAILJET_API_KEY_PRIVATE=<mailjet private api key>
```

**Note:** Only one mail service should be enabled at a time. The service priority is: Noop → Mailjet → SendGrid.

## Run

### Docker

For plain docker run the following commands:

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
curl -H "Content-Type: application/json" --data '{"subject":"my subject", "content":"my message", "to":"test@mail.com,test2@mail.com"}' http://localhost:80/v1/sendmail
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
