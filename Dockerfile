FROM golang:1.22-alpine as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go vet -v ./... 
RUN go test -v ./...

RUN CGO_ENABLED=0 go build -o /go/bin/app
# create empty folder to copy later into distoless image
RUN mkdir /secrets

FROM gcr.io/distroless/static-debian11

COPY --from=build /go/bin/app /
COPY --from=build /secrets /run/secrets

CMD ["/app"]