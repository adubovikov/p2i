NAME?=p2i

all:
	export PKG_CONFIG_PATH="`pwd`"; \
	CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-s -w" -o $(NAME)

debug:
	go build -o $(NAME)

modules:
	go get ./...


.PHONY: clean
clean:
	rm -fr $(NAME)
