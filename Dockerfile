FROM golang:1.3
COPY . /go/src/bitbucket.org/jacobgreenleaf/yeti
WORKDIR /go/src/bitbucket.org/jacobgreenleaf/yeti
RUN ./build.sh
CMD /go/bin/yeti
