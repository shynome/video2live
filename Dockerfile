FROM alpine:latest
RUN   apk add --no-cache ca-certificates ffmpeg
WORKDIR /app
VOLUME [ "/app/pb_data/" ]
EXPOSE 8090

COPY video2live /app/video2live
# start PocketBase
ENTRYPOINT [ "/app/video2live", "serve", "--http=0.0.0.0:8090" ]
CMD []
