SUB_DIR = server core config

.PHONY: build subdirs $(SUB_DIR)

subdirs: $(SUB_DIR)

$(SUB_DIR):
	@+make build-go -C $@

build-go:
	make subdirs
