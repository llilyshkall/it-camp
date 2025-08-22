FROM golang:1.24.5 as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o remarks ./cmd/remarks

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/remarks .
RUN chmod +x ./remarks
CMD ["./remarks"]
