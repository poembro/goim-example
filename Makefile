# Go parameters
GOCMD=GO111MODULE=on go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test

all: test build
build:
	rm -rf target/
	mkdir target/
	cp cmd/comet/comet-example.toml target/comet.toml
	cp cmd/logic/logic-example.toml target/logic.toml
	cp cmd/job/job-example.toml target/job.toml
	$(GOBUILD) -o target/comet cmd/comet/main.go 
	$(GOBUILD) -o target/logic cmd/logic/main.go
	$(GOBUILD) -o target/job cmd/job/main.go

test:
	$(GOTEST) -v ./...

clean:
	rm -rf target/

run:
	export GODEBUG=http2debug=2 && export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/job -conf=target/job.toml -region=sh -zone=sh001 -deploy.env=prod -weight=10 -addrs=47.111.69.116 -debug=true -host=127.0.0.1 -log_dir=./target -alsologtostderr
	export GODEBUG=http2debug=2 && export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/comet -conf=target/comet.toml -region=sh -zone=sh001 -deploy.env=prod -weight=10 -addrs=47.111.69.116 -debug=true -host=127.0.0.1 -log_dir=./target -alsologtostderr
	export GODEBUG=http2debug=2 && export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/logic -conf=target/logic.toml -region=sh -zone=sh001 -deploy.env=prod -weight=10 -host=127.0.0.1 -log_dir=./target -alsologtostderr

runlogic:
	export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/logic -conf=target/logic.toml -region=sh -zone=sh001 -deploy.env=prod -weight=10 -host=192.168.84.168 -log_dir=./target  -v=1 -alsologtostderr

runcomet:
	export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/comet -conf=target/comet.toml -region=sh -zone=sh001 -deploy.env=prod -weight=10 -addrs=47.111.69.116 -debug=true -host=192.168.84.168 -log_dir=./target  -v=1 -alsologtostderr

runjob:
	export GODEBUG=http2debug=2 && export GOLANG_PROTOBUF_REGISTRATION_CONFLICT=warn && ./target/job -conf=target/job.toml -region=sh -zone=sh001 -deploy.env=prod -host=192.168.84.168 -log_dir=./target  -v=1 -alsologtostderr

stop:
	pkill -f target/comet
	pkill -f target/logic
	pkill -f target/job
