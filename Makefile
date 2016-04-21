all: build

chimpanzee: 
	go build -o chimpanzee cmd/main.go

build: chimpanzee

clean:
	rm -f chimpanzee
