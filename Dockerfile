FROM golang:1.10 AS build
WORKDIR $GOPATH/src/github.com/dotnews/conduit
RUN curl -L https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 > /usr/bin/dep
RUN chmod +x /usr/bin/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app .

FROM alpine
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app ./
CMD ["/app", "-logtostderr=true"]
