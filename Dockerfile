FROM golang:1.25 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service ./cmd/service-courier
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o worker ./cmd/worker

FROM gcr.io/distroless/base-debian12 as service
WORKDIR /
COPY --from=builder /app/service /service-courier
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/service-courier"]

FROM gcr.io/distroless/base-debian12 as worker
WORKDIR /
COPY --from=builder /app/worker /worker
USER nonroot:nonroot
ENTRYPOINT ["/worker"]