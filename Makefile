build:
	podman run --rm -it -e BASHLY_TAB_INDENT=1 -v "${PWD}:/app" docker.io/dannyben/bashly generate --upgrade
go:
	go build -o ${HOME}/.local/bin/odago ./cmd/oda/.
	sudo mv ${HOME}/.local/bin/odago /usr/local/bin/oda
node:
	go build -o ${HOME}/.local/bin/odanode ./cmd/odanode/.

odai:
	@rm -f odai
	@go build -o odai ./cmd/odai/.