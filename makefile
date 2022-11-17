CMD = ./cmd/snash

GO = go

compile:
	$(GO) build -o ./bin/app $(CMD)

./bin:
	mkdir -p bin

run:
	$(GO) run $(CMD)

clean:
	rm -r ./bin/*

all:
	@echo compile, run, clean
