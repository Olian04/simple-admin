# simple-admin UI server.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Generate templ components, then build a static binary.
RUN go tool templ generate
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w" \
    -o /out/simple-admin ./cmd/simple-admin

FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/simple-admin /app/simple-admin
USER app
EXPOSE 8080
ENTRYPOINT ["/app/simple-admin"]
