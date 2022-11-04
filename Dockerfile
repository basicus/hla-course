FROM golang:alpine  AS build
LABEL stage=build_users

WORKDIR /src
COPY . .
RUN go install ./...

FROM alpine:3.14.6
COPY --from=build /go/bin/cmd /srv/cmd

EXPOSE 8080 8082 8084

ENTRYPOINT ["/srv/cmd"]
