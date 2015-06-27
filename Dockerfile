FROM golang
MAINTAINER Anders Pearson <anders@columbia.edu>
RUN apt-get update && apt-get install -y ca-certificates
RUN go get github.com/russross/blackfriday
RUN go get github.com/stvp/go-toml-config
RUN go get github.com/lib/pq
RUN go get github.com/nu7hatch/gouuid
ADD . /go/src/github.com/thraxil/gori
RUN go install github.com/thraxil/gori
RUN mkdir /gori/
EXPOSE 8890
ENV GORI_MEDIA_DIR=/go/src/github.com/thraxil/gori/media/
ENV GORI_PORT=8890
CMD ["/go/bin/gori"]

