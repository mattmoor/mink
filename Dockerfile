FROM alpine

ARG BUILD_DATE
ARG VERSION
ARG REVISION
ARG TARGETARCH
ARG TARGETOS

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add \
    ca-certificates curl git make netcat-openbsd
    
RUN echo using jx-mink version $VERSION and OS $TARGETOS arch $TARGETARCH && \
  cd /tmp && \
  curl -k -L https://github.com/jenkins-x-plugins/jx-mink/releases/download/v$VERSION/jx-mink-$TARGETOS-$TARGETARCH.tar.gz | tar xzv && \
  mv jx-mink /jx-mink

FROM gcr.io/kaniko-project/executor:debug-v1.3.0

ARG BUILD_DATE
ARG VERSION
ARG REVISION
ARG TARGETARCH
ARG TARGETOS

LABEL maintainer="jenkins-x"

COPY --from=0 /jx-mink /usr/bin/jx-mink

ENV HOME /kaniko
ENV PATH /usr/local/bin:/bin:/usr/bin:/kaniko:/ko-app

