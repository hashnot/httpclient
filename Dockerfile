FROM alpine:3.4
MAINTAINER Rafał Krupiński
ENV GOPATH=/go
ENTRYPOINT ["/go/bin/httpclient"]

ADD . /go/src/github.com/hashnot/httpclient
RUN cd /go/src/github.com/hashnot/httpclient &&\
    apk add --no-cache --update go git ca-certificates &&\
    go get &&\
    go build -o $GOPATH/bin/httpclient github.com/hashnot/httpclient &&\
    rm -rf /go/src &&\
    apk del --no-cache git go
