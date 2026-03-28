# ---- Builder ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /batch ./cmd/batch

# ---- Server Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot AS server

COPY --from=builder /server /server
COPY --from=builder /app/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/server"]

# ---- Batch Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot AS batch

COPY --from=builder /batch /batch
COPY --from=builder /app/migrations /migrations

ENTRYPOINT ["/batch"]
