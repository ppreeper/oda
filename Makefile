default: dev

install:
	@rm -f $HOME/go/bin/oda && go generate . > commit.txt && go install .
build:
	@go generate . > commit.txt
	@CGO_ENABLED=0 GOOS=linux go build -a -o ./bin/oda .
dev:
	@go generate . > commit.txt
	@go build -o ./bin/oda .
