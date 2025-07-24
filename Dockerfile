FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags "-s -w" -o /app/server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
ENTRYPOINT ["./server"]
