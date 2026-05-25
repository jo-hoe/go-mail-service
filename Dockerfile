FROM golang:1.26.3-alpine3.23 AS build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v ./... 
RUN go test -v ./...

RUN CGO_ENABLED=0 go build -o /go/bin/app ./internal/app
# create empty folder to copy later into distoless image
RUN mkdir /secrets

FROM gcr.io/distroless/static-debian12

# CONFIG_PATH is the only env var the app reads.
# Mount the config file at /config/config.yaml (see charts/go-mail-service or local/config.yaml).
ENV CONFIG_PATH=/config/config.yaml

COPY --from=build /go/bin/app /
COPY --from=build /secrets /run/secrets

CMD ["/app"]
