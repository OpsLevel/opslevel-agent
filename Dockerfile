FROM alpine:latest
COPY opslevel-agent /opslevel-agent
ENTRYPOINT ["/opslevel-agent"]
