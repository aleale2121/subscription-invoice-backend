# Build stage
FROM golang:1.19-alpine3.16 AS builder
WORKDIR /app
COPY go.mod go.sum ./ 
RUN go mod download
COPY . .
RUN  CGO_ENABLED=0 go build -o app ./cmd/api/ 

# Run stage
FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/app .
COPY assets /assets
COPY temp /temp
CMD [ "/app/app" ]