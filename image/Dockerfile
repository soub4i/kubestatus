

FROM golang:1.19-buster AS builder
LABEL maintainer="Abderrahim Soubai @soub4i"
ENV GO111MODULE=on
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY *.go ./
COPY index.html /index.html
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /kubestatus_http_server

FROM gcr.io/distroless/base-debian10
WORKDIR /
COPY --from=builder /kubestatus_http_server /kubestatus_http_server
COPY --from=builder /index.html /index.html
EXPOSE 8080
ENTRYPOINT ["/kubestatus_http_server"]
