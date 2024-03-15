FROM golang:1.21.3

WORKDIR /go/src
ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLED=1

EXPOSE 8082

RUN apt-get update

CMD ["tail", "-f", "/dev/null"]
