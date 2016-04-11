package main

import (
	"fmt"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/gocql/gocql"
)

// ./chimpanzee -u username -p password -c 1.2.3.4 -c 1.2.3.5 pcap 1.pcap 2.pcap 3.pcap

var (
	cassandraUsername, cassandraPassword *string
	cassandraHosts *[]string
)

func main() {
	chimpanzee := cli.App("chimpanzee", "Write data to netbrane defined cassandra tables")
	chimpanzee.Version("v version", "0.0.1")

	cassandraUsername = chimpanzee.StringOpt("u username", "", "Cassandra username")
	cassandraPassword = chimpanzee.StringOpt("p password", "", "Cassandra password")
	cassandraHosts = chimpanzee.StringsOpt("c cassandra_host", nil, "Cassandra host IPs")

	chimpanzee.Command("pcap", "Write a pcap file", WritePCAP)

	chimpanzee.Run(os.Args)
}

func OpenCassandraSession() (*gocql.Session, error) {
	cluster := gocql.NewCluster(*cassandraHosts...)
	cluster.Consistency = gocql.LocalOne
	cluster.ProtoVersion = 4
	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{10}
	cluster.Authenticator = gocql.PasswordAuthenticator{Username: *cassandraUsername, Password: *cassandraPassword}
	cluster.NumConns = 16

	cqlSession, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return cqlSession, nil
}
