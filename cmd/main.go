package main

import (
	"os"

	cli "github.com/jawher/mow.cli"
)

// ./chimpanzee -u username -p password -c 1.2.3.4 -c 1.2.3.5 pcap 1.pcap 2.pcap 3.pcap

func main() {
	chimpanzee := cli.App("chimpanzee", "Write data to netbrane defined cassandra tables")
	chimpanzee.Version("v version", "0.0.1")

	chimpanzee.Command("cassandra", "Write a netbrane protobuf file to cassandra", WriteCassandra)
	chimpanzee.Command("pcap_proto", "Write a pcap file", WritePCAPProtobuf)

	chimpanzee.Run(os.Args)
}
