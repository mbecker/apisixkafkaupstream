FROM --platform=linux/amd64 golang:1.18.3-alpine AS base

RUN apk -U add ca-certificates
RUN apk update && apk upgrade && apk add pkgconf git bash build-base sudo
RUN git clone https://github.com/edenhill/librdkafka.git && cd librdkafka && ./configure --prefix /usr && make && make install

FROM base as builder

ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -tags musl --ldflags "-extldflags -static" -o kafkaupstream .


# APISIX with JavaScript Plugin Runner
FROM apache/apisix:2.14.1-alpine

# APISIX Go Plugin Runner
COPY --from=builder /build/kafkaupstream /usr/local/apisix/plugins/kafkaupstream