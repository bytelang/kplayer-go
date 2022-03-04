FROM alpine:latest
COPY build/kplayer /usr/bin/kplayer
WORKDIR /kplayer
ENTRYPOINT ["/usr/bin/kplayer"]