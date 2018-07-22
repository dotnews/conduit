FROM golang:1.10 AS build
WORKDIR $GOPATH/src/github.com/dotnews/conduit
RUN curl -L https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 > /usr/bin/dep
RUN chmod +x /usr/bin/dep
ADD https://api.github.com/repos/dotnews/crawler/compare/master...HEAD /dev/null
RUN git clone https://github.com/dotnews/crawler.git /crawler
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /app .

FROM node:alpine
RUN apk add --no-cache udev ttf-freefont chromium
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /crawler /crawler
COPY --from=build /app /app
RUN cd /crawler && npm install && cd /
CMD ["/app", "-logtostderr=true"]
