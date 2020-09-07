NAME?=request

all:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w"  -o bin/$(NAME)-darwin $(NAME).go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"  -o bin/$(NAME)-linux $(NAME).go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w"  -o bin/$(NAME).exe $(NAME).go

debug:
	go build -o $(NAME) *.go

.PHONY: clean
clean:
	rm -fr bin
