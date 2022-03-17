FROM alpine:latest
COPY build/kplayer /usr/bin/kplayer
WORKDIR /kplayer
CMD ["/usr/bin/kplayer", "play", "start"]