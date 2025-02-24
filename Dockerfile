FROM golang:alpine as build
RUN apk --no-cache add ca-certificates

FROM alpine:latest
# copy the ca-certificate.crt from the build stage
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY opslevel-agent /opslevel-agent
ENTRYPOINT ["/opslevel-agent"]
