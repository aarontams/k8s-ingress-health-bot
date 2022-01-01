FROM golang:1.17.5-alpine AS builder

ARG BUILDVERSION
ARG BUILDDATE

WORKDIR /app
COPY . .
ENV CGO_ENABLED 0
RUN go mod vendor; go build \
    -ldflags '-extldflags "-static"' \
    -ldflags "-X main.buildVersion=${BUILDVERSION} -X main.buildDate=${BUILDDATE}" \
    -o k8s-ingress-health-bot ./cmd/k8s-ingress-health-bot

FROM scratch

COPY --from=builder /app/k8s-ingress-health-bot /usr/bin/
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/

ENTRYPOINT ["k8s-ingress-health-bot"]

EXPOSE 8088
