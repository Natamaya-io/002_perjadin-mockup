# Stage 1: Build the Go binary
FROM public.ecr.aws/docker/library/golang:1.25-alpine AS builder

# Install SSL ca certificates and UPX
RUN apk --no-cache add ca-certificates tzdata upx

WORKDIR /app

# Copy Go modules first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -extldflags '-static'" -o main ./cmd/api/main.go

# Stage 2: Ultra-lightweight runtime
FROM public.ecr.aws/docker/library/alpine:latest

# Install tzdata for proper time handling in Go
RUN apk --no-cache add tzdata ca-certificates

WORKDIR /app

# Copy the compiled, stripped binary
COPY --from=builder /app/main .

# Copy migrations
COPY db/migrations ./db/migrations

# Copy DOCX templates folder
COPY templates ./templates

# Go runtime memory optimization (Zero-Trust & Anti-OOM)
ENV GOMEMLIMIT="2500MiB"
ENV GOGC=50

# Expose the application port
EXPOSE 8081

# Command to run the application
ENTRYPOINT ["./main"]

