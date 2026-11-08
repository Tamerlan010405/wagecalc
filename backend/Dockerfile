FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /out/wagecalc-api ./cmd/api

FROM alpine:3.20

WORKDIR /app

COPY --from=build /out/wagecalc-api /app/wagecalc-api

EXPOSE 8080

CMD ["/app/wagecalc-api"]
