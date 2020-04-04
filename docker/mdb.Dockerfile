FROM golang:1.13.7-stretch

ENV GO111MODULE=on

WORKDIR $GOPATH/src/github.com/pion/ion

COPY go.mod go.sum ./
RUN apt-get update && apt-get install -y --no-install-recommends libssl-dev
RUN cd $GOPATH/src/github.com/pion/ion && go mod download

COPY . $GOPATH/src/github.com/pion/ion

WORKDIR $GOPATH/src/github.com/pion/ion/pkg/node/mdb
RUN CGO_ENABLED=1 GOOS=linux go build -a -i -o /mdb .

FROM alpine:3.9.5

RUN apk --no-cache add ca-certificates
COPY --from=0 /mdb /usr/local/bin/mdb

ENTRYPOINT ["/usr/local/bin/mdb"]
