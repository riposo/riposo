FROM golang:1.16.2-buster

COPY build-env/build-plugin.sh /usr/local/bin/
ENTRYPOINT [ "build-plugin.sh" ]

COPY . /riposo
RUN set -eux; \
  \
  (cd /riposo; go mod download)
