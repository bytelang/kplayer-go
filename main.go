package main

// #cgo pkg-config: spdlog libavcodec libavutil libavformat libavfilter fmt
// #cgo CFLAGS: -I/Users/kangkai/smart/develop/cpp/libkplayer
// #cgo LDFLAGS: -L/Users/kangkai/smart/develop/cpp/libkplayer/build -L/Users/kangkai/smart/develop/cpp/libkplayer/build/util -L/Users/kangkai/smart/develop/cpp/libkplayer/build/codec -lkplayer -lkpcodec -lkputil -lstdc++
// #include "kplayer.h"
import "C"

func main() {
    C.Run()
}
