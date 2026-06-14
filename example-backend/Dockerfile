# Reference compliant backend (PocketBase). Separate module → built standalone.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# modernc sqlite is pure Go → CGO can stay disabled for a static binary.
RUN CGO_ENABLED=0 go build -trimpath -o /out/example-backend .

FROM alpine:3.20
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/example-backend /app/example-backend
USER app
EXPOSE 8090
ENTRYPOINT ["/app/example-backend"]
CMD ["serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
