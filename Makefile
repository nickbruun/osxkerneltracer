PREFIX ?= /usr/local

all: osxkerneltracer

osxkerneltracer: main.go
	go build -o osxkerneltracer .

install: osxkerneltracer
	install osxkerneltracer "$(PREFIX)/bin/osxkerneltracer"
