FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o captioner .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/captioner .
# Arial Unicode.ttf is macOS-licensed; mount at runtime via -v /Library/Fonts:/sysfonts:ro
# and set FONT_PATH, or COPY your own font to /app/font.ttf and set FONT_PATH=/app/font.ttf
ENV FONT_PATH=/Library/Fonts/Arial Unicode.ttf
ENTRYPOINT ["./captioner"]
