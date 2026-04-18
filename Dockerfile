FROM golang:1.25-alpine AS build
ARG VERSION=dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${VERSION}" -o /machine-agent ./cmd/machine-agent

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /machine-agent /machine-agent
COPY --from=build /etc/passwd /etc/passwd
USER nobody
EXPOSE 7891
ENTRYPOINT ["/machine-agent"]
