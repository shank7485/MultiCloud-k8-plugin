FROM alpine:3.7

ARG HTTP_PROXY=${HTTP_PROXY}
ARG HTTPS_PROXY=${HTTPS_PROXY}

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apk update 
# RUN apk add --no-cache bash

EXPOSE 8081

WORKDIR /plugin
ADD ./k8plugin /plugin/
ADD ./entrypoint.sh /plugin/

ENTRYPOINT /plugin/entrypoint.sh