build-go:
	protoc -I=.:${PROTO_PATH} --go_out=. --go-grpc_out=. --grpc-gateway_out=logtostderr=true:. --openapiv2_out=./gen *.proto && \
    protoc -I=.:${PROTO_PATH} --grpc-gateway_out=. *.proto
	cp -rf github.com/bytelang/kplayer/types/* ../../types/
	rm -rf github.com

clear:
	rm -rf *.go

