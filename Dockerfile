# Build stage - Frontend
FROM node:20-alpine AS frontend
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Build stage - Backend
FROM golang:1.22-alpine AS backend
WORKDIR /app
RUN apk --no-cache add git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/web/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o apisto ./cmd/apisto/

# Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend /app/apisto .
COPY --from=frontend /app/web/dist ./web/dist
EXPOSE 8080
CMD ["./apisto"]
