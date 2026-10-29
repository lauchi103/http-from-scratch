FROM ubuntu:latest
LABEL authors="sharifi"

ENTRYPOINT ["top", "-b"]