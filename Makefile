build:
	BASHLY_TAB_INDENT=1  podman run --rm -it --volume "${PWD}:/app" docker.io/dannyben/bashly generate