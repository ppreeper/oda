build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a .