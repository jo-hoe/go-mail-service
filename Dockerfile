FROM golang:1.24.2-alpine3.20 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v ./... 
RUN go test -v ./...

RUN CGO_ENABLED=0 go build -o /go/bin/app
# create empty folder to copy later into distoless image
RUN mkdir /secrets

FROM gcr.io/distroless/static-debian11

ENV API_PORT=80 \
    IS_NOOP_ENABLED=true \
    IS_SENDGRID_ENABLED=false

COPY --from=build /go/bin/app /
COPY --from=build /secrets /run/secrets

CMD ["/app"]
