NAME?=p2i

all:
	CGO_ENABLED=1 GOOS=linux go build -a -ldflags "-s -w" -o $(NAME)

debug:
	go build -o $(NAME)

modules:
	go get ./...


.PHONY: clean
clean:
	rm -fr $(NAME)
