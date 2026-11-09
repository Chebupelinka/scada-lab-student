FROM golang:1.26.10-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /gateway .

FROM alpine:3.23
COPY --from=build /gateway /usr/local/bin/gateway
ENTRYPOINT ["gateway"]
