FROM golang:1.3
COPY . /go/src/github.com/jacobgreenleaf/yeti
WORKDIR /go/src/github.com/jacobgreenleaf/yeti
RUN ./build.sh
CMD /go/bin/yeti
