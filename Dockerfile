FROM gcr.io/distroless/base-debian11:nonroot
COPY riposo /usr/local/bin/
ENTRYPOINT ["riposo"]
EXPOSE 8888
