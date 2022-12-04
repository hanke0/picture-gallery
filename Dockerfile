from golang:1.19 as builder


WORKDIR /app

COPY . /app

RUN go build -o picture-gallery .

from alpine:latest

WORKDIR /

COPY --from=build /app/picture-gallery /app/picture-gallery

RUN apk update && apk install curl && chmod +x picture-gallery

EXPOSE 9288

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD ["curl","--connect-timeout","1","127.0.0.1:9288"]
CMD ["/app/picture-gallery"]
