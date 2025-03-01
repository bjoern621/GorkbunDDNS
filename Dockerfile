# syntax=docker/dockerfile:1
# check=skip=SecretsUsedInArgOrEnv

FROM golang:1.24.0 AS go-build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN go build -o start-gorkbunddns

# Go build complete

FROM gcr.io/distroless/base-debian12
WORKDIR /app

COPY --from=go-build /app/start-gorkbunddns ./

# Copied binary to distroless image

CMD [ "/app/start-gorkbunddns" ]

ENV  DOMAINS=example.com,sub.example.com,*.example.com APIKEY=pk1_xyz SECRETKEY=sk1_xyz TIMEOUT=600