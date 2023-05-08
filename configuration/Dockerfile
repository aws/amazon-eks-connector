# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build
ARG TARGETOS TARGETARCH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-s -w" -o bin/connector ./cmd

FROM public.ecr.aws/amazonlinux/amazonlinux:2 AS cert

FROM scratch
COPY --from=build /app/bin/connector /var/eks/connector
# copy ca-bundle.crt from AmazonLinux2...
COPY --from=cert /etc/ssl/certs/ca-bundle.crt /etc/ssl/certs/
ENTRYPOINT ["/var/eks/connector"]
