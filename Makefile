oda:
	@rm -f $HOME/go/bin/oda && go install ./cmd/oda/.
odabuild:
	@CGO_ENABLED=0 GOOS=linux go build -a -o bin/oda ./cmd/oda/.
odaserver:
	@rm -f $HOME/go/bin/odaserver && go install ./cmd/odaserver/.
odaserverbuild:
	@CGO_ENABLED=0 GOOS=linux go build -a -o bin/odaserver ./cmd/odaserver/.
build:
	@CGO_ENABLED=0 GOOS=linux go build -o bin/odaserver ./cmd/odaserver/.