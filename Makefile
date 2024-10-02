oda:
	@rm -f $HOME/go/bin/oda && go generate ./cmd/oda/. > ./cmd/oda/commit.txt && go install ./cmd/oda/.
odabuild:
	@go generate ./cmd/oda/. > ./cmd/oda/commit.txt
	@CGO_ENABLED=0 GOOS=linux go build -a -o bin/oda ./cmd/oda/.
odas:
	@rm -f $HOME/go/bin/odas && go generate ./cmd/odas/. > ./cmd/odas/commit.txt &&  go install ./cmd/odas/.
odasbuild:
	@go generate ./cmd/odas/. > ./cmd/odas/commit.txt
	@CGO_ENABLED=0 GOOS=linux go build -a -o bin/odas ./cmd/odas/.
