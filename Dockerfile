FROM alpine:3.4

RUN apk add --no-cache ca-certificates

ADD ./draughtsman-operator /draughtsman-operator

ENTRYPOINT ["/draughtsman-operator"]
