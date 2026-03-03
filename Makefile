PREFIX ?= $(HOME)/.local

build:
	go build -o mdp .

install: build
	mkdir -p $(PREFIX)/bin
	cp mdp $(PREFIX)/bin/mdp

uninstall:
	rm -f $(PREFIX)/bin/mdp

clean:
	rm -f mdp

.PHONY: build install uninstall clean
