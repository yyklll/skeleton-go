FROM golang:1.13.7-alpine3.11 as builder

ENV GOPROXY https://goproxy.cn

RUN mkdir -p /skeleton/

WORKDIR /skeleton

COPY . .

RUN go mod download

RUN go test -v -race ./...

RUN GIT_COMMIT=$(git rev-list -1 HEAD) && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w \
    -X main.buildstamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` -X main.githash=${GIT_COMMIT}" \
    -a -o bin/server cmd/skeleton/*

FROM alpine:3.10

RUN apk install dumb-init

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    curl openssl netcat-openbsd

WORKDIR /home/app

COPY --from=builder /skeleton/bin/server .
RUN chown -R app:app ./

USER app

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["./server"]
