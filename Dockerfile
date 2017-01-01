FROM alpine:3.4
MAINTAINER Rafał Krupiński
ENV GOPATH=/go
EXPOSE 8080
ENTRYPOINT ["/go/bin/httpclient"]

ADD . /go/src/bitbucket.org/hashnot/httpclient
RUN cd /go/src/bitbucket.org/hashnot/httpclient &&\
    apk add --no-cache --update go git &&\
    go get &&\
    go build -o $GOPATH/bin/httpclient bitbucket.org/hashnot/httpclient &&\
    apk del --no-cache git go
