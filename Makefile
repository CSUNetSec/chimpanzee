all: build

chimpanzee: 
	go build -o chimpanzee cmd/*

build: chimpanzee

clean:
	rm -f chimpanzee
