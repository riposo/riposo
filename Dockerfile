FROM gcr.io/distroless/base-debian10:nonroot
COPY riposo /usr/local/bin/
ENTRYPOINT ["riposo"]
EXPOSE 8888
