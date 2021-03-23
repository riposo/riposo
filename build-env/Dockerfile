FROM golang:1.16.2-buster@sha256:a23a7e49a820f9ae69df0fedf64f037cb15b004997effa93ec885e5032277bc1

COPY build-env/build-plugin.sh /usr/local/bin/
ENTRYPOINT [ "build-plugin.sh" ]

COPY . /riposo
RUN set -eux; \
  \
  (cd /riposo; go mod download)
