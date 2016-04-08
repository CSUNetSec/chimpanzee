package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
)

// ./chimpanzee -uusername -ppassword -h1.2.3.4 -h1.2.3.5 pcap 1.pcap 2.pcap 3.pcap

func main() {
	chimpanzee := cli.App("chimpanzee", "Write data to netbrane defined cassandra tables")
	chimpanzee.Version("v version", "0.0.1")

	cassandraUsername := chimpanzee.StringOpt("u username", "", "Cassandra username")
	cassandraPassword := chimpanzee.StringOpt("p password", "", "Cassandra password")
	cassandraHosts := chimpanzee.StringsOpt("h host", nil, "Cassandra host IPs")

	chimpanzee.Command("pcap", "Write a pcap file", func(cmd *cli.Cmd) {
		cmd.Spec = "FILENAME..."
		filenames := cmd.StringsArg("FILENAME", nil, "pcap files to write")

		cmd.Action = func() {
			for _, filename := range *filenames {
				fmt.Printf("writing pcap file '%s' to '%v' username:%s password:%s\n", filename, *cassandraHosts, *cassandraUsername, *cassandraPassword)
			}
		}
	})

	chimpanzee.Run(os.Args)
}
