build:
	podman run --rm -it -e BASHLY_TAB_INDENT=1 -v "${PWD}:/app" docker.io/dannyben/bashly generate
