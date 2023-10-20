ARG ALPINE_TAG=3.16
ARG GO_TAG=1.19.2-alpine3.16

FROM golang:${GO_TAG} AS builder

WORKDIR /app

COPY go.mod .
COPY go.sum .
COPY main.go .

RUN go get ./...
RUN go build -o main ./...


FROM alpine:${ALPINE_TAG}

WORKDIR /app

RUN apk update && apk --no-cache add curl

COPY --from=builder /app/main main

COPY fortune.txt .
COPY static static
COPY views views

ENV PORT=3000 GIN_MODE=release

EXPOSE ${PORT}

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
	CMD curl -s http://localhost:${PORT}/healthz || exit 1

ENTRYPOINT [ "./main" ]

CMD [ "" ]
