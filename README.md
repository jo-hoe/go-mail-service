# Mail Service

[![Test Status](https://github.com/jo-hoe/go-mail-service/workflows/test/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)
[![Go Report Card](https://goreportcard.com/badge/github.com/jo-hoe/go-mail-service)](https://goreportcard.com/report/github.com/jo-hoe/go-mail-service)
[![Coverage Status](https://coveralls.io/repos/github/jo-hoe/go-mail-service/badge.svg?branch=main)](https://coveralls.io/github/jo-hoe/go-mail-service?branch=main)

## Example Request

```bash
curl -H "Content-Type: application/json" --data '{"subject":"my subject", "content":"my message", "to":["test@mail.de"]}' http://localhost:80/v1/sendmail
```
