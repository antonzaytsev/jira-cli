# Build:   docker build -t jira-cli .
# Run:     docker run --rm jira-cli version
# Extract: docker build --output=type=local,dest=./bin --target=export .

FROM golang:1.25-alpine3.23 AS builder

ARG TARGETOS
ARG TARGETARCH

ENV CGO_ENABLED=0

WORKDIR /app

COPY . .

RUN apk add -U --no-cache make git && make deps install

FROM builder AS export
RUN cp /go/bin/jira /jira

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/jira /usr/local/bin/jira
ENTRYPOINT ["jira"]
