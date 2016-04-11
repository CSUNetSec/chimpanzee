all: build

chimpanzee: 
	go build -o chimpanzee cmd/* || (echo "running go get";cd cmd/; go get; go get -u;cd ../; go build -o chimpanzee cmd/*);\

build: chimpanzee

clean:
	rm -f chimpanzee
