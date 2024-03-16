# Mail Service

[![Test Status](https://github.com/jo-hoe/go-mail-service/workflows/test/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=test)
[![Lint Status](https://github.com/jo-hoe/go-mail-service/workflows/lint/badge.svg)](https://github.com/jo-hoe/go-mail-service/actions?workflow=lint)

## Example Request

```bash
curl -H "Content-Type: application/json" --data '{"subject":"my subject", "content":"my message", "to":["test@mail.de"]}' http://localhost:80/v1/sendmail
```
