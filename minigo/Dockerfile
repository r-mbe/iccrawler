FROM frolvlad/alpine-glibc:alpine-3.6
MAINTAINER stan <stanhangzhou@qq.com>


RUN apk update && apk upgrade && \

    # Make info file about this build
    apk add ca-certificates &&\
    mkdir -p /etc/BUILDS/ && \
    mkdir -p /usr/local && \
    printf "Build of nimmis/alpine-golang:1.8.3, date: %s\n"  `date -u +"%Y-%m-%dT%H:%M:%SZ"` > /etc/BUILDS/alpine-golang && \

    # add extra package for installation
    apk add curl && \

    cd /&& \

    # download and upack golang
    curl https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz | tar xzf - -C /usr/local/ && \

    # remove extra packages
    apk del curl && \

    #mkidr gopath directory
    mkdir -p /go /go/bin /go/src /go/pkg && \

    # remove cached info
    rm -rf /var/cache/apk/*


ENV GOROOT=/usr/local/go \
    GOPATH=/go \
    GOBIN=/go/bin \
    PATH=${PATH}:${GOROOT}/bin:/${GOPATH}/bin:${GOBIN}

ADD . ${GOPATH}/src/app/
WORKDIR ${GOPATH}/src/app
#COPY . .

#RUN go get github.com/gorilla/mux
RUN /usr/local/go/bin/go build -o main .

CMD ["/go/src/app/main"]
