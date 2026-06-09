FROM golang:1.26-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ENV CGO_ENABLED=1
RUN go build -ldflags="-s -w" -trimpath -buildvcs=false -o main .

FROM alpine:latest
RUN apk --no-cache add sqlite-libs
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/fonts ./fonts
EXPOSE 8080
CMD ["./main"]
