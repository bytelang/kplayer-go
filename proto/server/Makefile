build-go:
	protoc -I=.:${PROTO_PATH} --gogo_out=plugins=grpc:. --grpc-gateway_out=logtostderr=true:. --openapiv2_out=./gen *.proto
	cp -rf github.com/bytelang/kplayer/types/* ../../types/
	rm -rf github.com

clear:
	rm -rf *.go

