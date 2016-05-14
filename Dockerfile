# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
# docker build  --tag="opencoredata/ocdSearch:0.1"  .
# docker run -d -p 6789:6789  opencoredata/ocdSearch:0.1
FROM golang:1.6

# Copy the local package files to the container's workspace.
ADD . /go/src/opencoredata.org/ocdSearch
#ADD https://codeload.github.com/OpenCoreData/ocdCommons/zip/master  / 

# Uncompress ocdCommons
# code here to uncompress and move the commons package


# Create a non-root user to run as
RUN groupadd -r gorunner -g 433 && \
mkdir /home/gorunner && \
useradd -u 431 -r -g gorunner -d /home/gorunner -s /sbin/nologin -c "User to run go apps on high ports" gorunner && \
chown -R gorunner:gorunner  /home/gorunner && \
chown -R gorunner:gorunner /go/src/opencoredata.org/ocdSearch


# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/gorilla/mux
RUN go get github.com/blevesearch/bleve
RUN go get github.com/parnurzeal/gorequest
RUN go get github.com/couchbase/moss
RUN go get github.com/syndtr/goleveldb/leveldb
RUN go get golang.org/x/text/unicode/norm

# set user
USER gorunner

# Move to a workign directory for running codices so it can see it's static files
# future version should take this as a param so static content can be anywhere
WORKDIR /go/src/opencoredata.org/ocdSearch
RUN go build .


# Run the command by default when the container starts.
ENTRYPOINT /go/src/opencoredata.org/ocdSearch/ocdSearch

# Document that the service listens on this port
# container needs to talk to database container
EXPOSE 9802
#EXPOSE 9802 9800
