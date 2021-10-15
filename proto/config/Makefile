build-go:
	protoc -I=. -I=${PROTO_PATH} --gogo_out=. *.proto
	cp -rf github.com/bytelang/kplayer/types/ ../../types/
	rm -rf github.com

clear:
	rm -rf *.go

