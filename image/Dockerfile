

FROM golang:1.16-buster AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
COPY index.html /index.html
RUN go build -o /kubestatus_http_server

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /kubestatus_http_server /kubestatus_http_server
COPY --from=builder /index.html /index.html
EXPOSE 8080
ENTRYPOINT ["/kubestatus_http_server"]
