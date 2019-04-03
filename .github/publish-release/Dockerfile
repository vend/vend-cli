FROM golang:1.12.1-alpine3.9

LABEL "com.github.actions.name"="Publish Release"
LABEL "com.github.actions.description"="Publishes a Github Release"
LABEL "com.github.actions.icon"="briefcase"
LABEL "com.github.actions.color"="purple"

RUN apk add --no-cache git openssh jq

ADD entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]