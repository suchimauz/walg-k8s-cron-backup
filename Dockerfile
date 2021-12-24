FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /var/data/app

COPY .bin/app app

CMD ["./app"]