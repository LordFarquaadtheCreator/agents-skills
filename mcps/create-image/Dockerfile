FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o create-image .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/create-image .
COPY --from=builder /app/model_card.yaml .
ENV COMFYUI_API_URL=""
ENTRYPOINT ["./create-image"]
