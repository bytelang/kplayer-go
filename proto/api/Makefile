build-go:
	protoc -I=.:${PROTO_PATH} --go_out=. *.proto
	cp -rf github.com/bytelang/kplayer/types/* ../../types/
	rm -rf github.com

clear:
	rm -rf *.go

