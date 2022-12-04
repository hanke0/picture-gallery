FROM golang:1.19-alpine as builder
COPY . /app/
RUN cd /app && go build -o picture-gallery ./cmd/main.go

FROM alpine:latest
WORKDIR /
RUN apk update && apk add curl && rm -rf /tmp/* && mkdir /pictures
COPY --from=builder /app/picture-gallery /picture-gallery
EXPOSE 9288
HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD ["curl","--connect-timeout","1","127.0.0.1:9288/ping"]
CMD ["/picture-gallery"]
