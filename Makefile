build:
	podman run --rm -it -e BASHLY_TAB_INDENT=1 -v "${PWD}:/app" docker.io/dannyben/bashly generate --upgrade
go:
	go build -o ${HOME}/.local/bin/odago
	sudo mv ${HOME}/.local/bin/odago /usr/local/bin/oda
