FROM --platform=$BUILDPLATFORM golang:1.26.1 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS=linux
ARG TARGETARCH=amd64

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/app .

FROM gcr.io/distroless/static-debian12:nonroot

ENV PORT=8080
ENV GIN_MODE=release

WORKDIR /

COPY --from=builder /out/app /app

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app"]
