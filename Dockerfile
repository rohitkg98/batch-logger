FROM golang:1.16 AS builder
WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o app

FROM golangci/golangci-lint:v1.39.0 AS linter

FROM builder AS lint
COPY --from=linter /usr/bin/golangci-lint /usr/bin/golangci-lint
RUN golangci-lint run

FROM golang:1.16
COPY --from=builder /src/app app
EXPOSE 8080
CMD ./app
