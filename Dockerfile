FROM gliderlabs/alpine:3.3

COPY . /go/src/github.com/bobrik/docker-image-cleaner

RUN apk-install go git && \
    GOPATH=/go go get github.com/bobrik/docker-image-cleaner && \
    apk del go git && \
    mv /go/bin/docker-image-cleaner /bin/docker-image-cleaner && \
    rm -rf /go

ENTRYPOINT ["/bin/docker-image-cleaner"]
