package converter

import (
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/gocql/gocql"
	pb "github.com/CSUNetSec/netbrane_proto/nofutz_FlowMinder"
)

const (
	pcapInsertStmt = "INSERT INTO netbrane_pcap_core.packets_by_time(time_bucket, capture_host, timestamp, packet_size, source_mac, destination_mac, ip_protocol, source_ip, destination_ip, ip_flags, source_port, destination_port, tcp_flags, tcp_window_size, tcp_sequence, tcp_acknowledgement) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
)

var timeBucketSize *int

func WriteCassandra(cmd *cli.Cmd) {
	cmd.Spec = "[-t] [-w] [-c] [-u] [-p] [-c] FILENAME..."

	timeBucketSize = cmd.IntOpt("t time_bucket", 600, "Number of seconds in a time bucket")
	workerCount := cmd.IntOpt("w worker_count", 100, "Number of insertion workers")
	connectionCount := cmd.IntOpt("c connection_count", 16, "Number of connections to cassandra from cql session")
	username := cmd.StringOpt("u username", "", "Cassandra username")
	password := cmd.StringOpt("p password", "", "Cassandra password")
	hosts := cmd.StringsOpt("c cassandra_hosts", nil, "Cassandra host IPs")
	filenames := cmd.StringsArg("FILENAME", nil, "Filenames to be loaded into cassandra")

	cmd.Action = func() {
		//open cassandra session
		cluster := gocql.NewCluster(*hosts...)
		cluster.Consistency = gocql.LocalOne
		cluster.ProtoVersion = 4
		cluster.RetryPolicy = &gocql.SimpleRetryPolicy{10}
		cluster.NumConns = *connectionCount

		if username != nil && password != nil {
			cluster.Authenticator = gocql.PasswordAuthenticator{Username: *username, Password: *password}
		}

		cqlSession, err := cluster.CreateSession()
		if err != nil {
			panic(err)
		}

		//loop through files
		for _, filename := range *filenames {
			_, err := os.Open(filename)
			if err != nil {
				panic(err)
			}

			workChan := make(chan *pb.CaptureRecordUnion)
			resultChan := make(chan bool)

			//start workers
			for i := 0; i < *workerCount; i++ {
				go worker(cqlSession, workChan, resultChan)
			}

			//read records from file and write to workChan
		}
	}
}

func worker(cqlSession *gocql.Session, workChan chan *pb.CaptureRecordUnion, resultChan chan bool) error {
	for {
		_ = <-workChan

		//TODO write protobuf to cql session

		resultChan <- true
	}

	return nil
}
