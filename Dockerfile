FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -o /out/workshop-server ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
USER appuser

COPY --from=build /out/workshop-server /usr/local/bin/workshop-server

EXPOSE 3000

ENTRYPOINT ["workshop-server"]
