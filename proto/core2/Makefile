build-cpp:
	protoc -I=. --cpp_out=. *.proto
	cd prompt && protoc -I=. --cpp_out=. *.proto
	cd msg && protoc -I=. --cpp_out=. *.proto

build-go:
	protoc -I=. -I=${PROTO_PATH} -I=./prompt -I=./msg --gogo_out=. *.proto
	protoc -I=. -I=${PROTO_PATH} -I=./prompt -I=./msg --gogo_out=. prompt/*.proto
	protoc -I=. -I=${PROTO_PATH} -I=./prompt -I=./msg --gogo_out=. msg/*.proto
	cp -rf github.com/bytelang/kplayer/types/* ../../types/
	rm -rf github.com

clear:
	rm -rf *.pb.cc *.h *.go
	rm -rf prompt/*.pb.cc prompt/*.h prompt/*.go
	rm -rf msg/*.pb.cc msg/*.h msg/*.go

