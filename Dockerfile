# Use a multi-stage build for Go
FROM golang:1.20 AS go_exec_builder
WORKDIR /app
COPY BibleServe.go .
COPY go.mod .
COPY go.sum .
RUN go build -o BibleServe BibleServe.go

FROM alpine:latest
RUN apk add --no-cache libc6-compat

COPY --from=go_exec_builder /app/BibleServe /app/BibleServe
COPY bible/ESVBible.txt /app/bible/ESVBible.txt

WORKDIR /app

CMD ["./BibleServe"]