FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /agent-runner-client cmd/agent-runner-client/main.go

FROM alpine:3.23
COPY --from=builder /agent-runner-client /bin/agent-runner-client
COPY build/agent-runner-client/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
ENTRYPOINT ["agent-runner-client"]
