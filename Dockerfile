FROM --platform=$BUILDPLATFORM golang:1.24 AS go-build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

ARG TARGETOS
ARG TARGETARCH

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o start-gorkbunddns

# Go build complete

FROM gcr.io/distroless/base-debian12 AS debian12
WORKDIR /app

COPY --from=go-build /app/start-gorkbunddns ./

# Copied binary to distroless image

CMD [ "/app/start-gorkbunddns" ]