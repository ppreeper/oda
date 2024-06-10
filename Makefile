oda:
	@rm -f $HOME/go/bin/oda && go install ./cmd/oda/.
odaserver:
	@rm -f $HOME/go/bin/odaserver && go install ./cmd/odaserver/.