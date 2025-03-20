FROM alpine:latest

ARG CI_PROJECT_NAME=apisrv
ENV CI_PROJECT_NAME=$CI_PROJECT_NAME

ENV TZ=Europe/Moscow
RUN apk --no-cache add ca-certificates tzdata && cp -r -f /usr/share/zoneinfo/$TZ /etc/localtime

ADD ./$CI_PROJECT_NAME ./$CI_PROJECT_NAME

LABEL description="$CI_PROJECT_NAME"

ENTRYPOINT "/$CI_PROJECT_NAME"
EXPOSE 8075
