package main

import (
	"fmt"
	"net"
	"os"
	"time"

	cli "github.com/jawher/mow.cli"
	"github.com/gocql/gocql"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// ./chimpanzee -uusername -ppassword -h1.2.3.4 -h1.2.3.5 pcap 1.pcap 2.pcap 3.pcap

const (
	pcapInsertStmt = "INSERT INTO netbrane_pcap_core.packets_by_time(time_bucket, capture_host, timestamp, packet_size, source_mac, destination_mac, ip_protocol, source_ip, destination_ip, ip_flags, source_port, destination_port, tcp_flags, tcp_window_size, tcp_sequence, tcp_acknowledgement) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
)

func main() {
	chimpanzee := cli.App("chimpanzee", "Write data to netbrane defined cassandra tables")
	chimpanzee.Version("v version", "0.0.1")

	cassandraUsername := chimpanzee.StringOpt("u username", "", "Cassandra username")
	cassandraPassword := chimpanzee.StringOpt("p password", "", "Cassandra password")
	cassandraHosts := chimpanzee.StringsOpt("c cassandra_host", nil, "Cassandra host IPs")

	//process pcap files
	chimpanzee.Command("pcap", "Write a pcap file", func(cmd *cli.Cmd) {
		cmd.Spec = "[-t] CAPTURE_HOST FILENAME..."
		timeBucketSize := int64(*cmd.IntOpt("t time_bucket", 600, "Number of seconds in a time bucket"))
		captureHostname := cmd.StringArg("CAPTURE_HOST", "", "Host the packets were captured from")
		filenames := cmd.StringsArg("FILENAME", nil, "Pcap files to write")

		cmd.Action = func() {
			//connect to cassandra
			cluster := gocql.NewCluster(*cassandraHosts...)
			cluster.Consistency = gocql.LocalOne
			cluster.ProtoVersion = 4
			cluster.RetryPolicy = &gocql.SimpleRetryPolicy{10}
			cluster.Authenticator = gocql.PasswordAuthenticator{Username: *cassandraUsername, Password: *cassandraPassword}
			cluster.NumConns = 16

			cqlSession, err := cluster.CreateSession()
			if err != nil {
				panic(err)
			}

			for _, filename := range *filenames {
				startTime := time.Now()
				fmt.Printf("writing pcap file '%s' to '%v' username:%s password:%s\n", filename, *cassandraHosts, *cassandraUsername, *cassandraPassword)

				//open pcap file
				handle, err := pcap.OpenOffline(filename)
				if err != nil {
					fmt.Printf("%s\n", err)
					continue
				}

				defer handle.Close()

				//declare packet variables
				var (
					srcMAC, dstMAC *net.HardwareAddr
					ipProtocol int
					srcIP, dstIP *net.IP
					ipFlags uint8
					srcPort, dstPort uint16
					sequence, acknowledgement uint32
					windowSize uint16
				)

				//loop through pcap packets
				packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
				for packet := range packetSource.Packets() {
					//parse link layer
					linkLayer := packet.LinkLayer()
					if linkLayer == nil {
						fmt.Printf("No link layer information, skipping packet\n")
						continue
					}

					switch linkLayer.(type) {
					case *layers.Ethernet:
						ethernetLayer, _ := linkLayer.(*layers.Ethernet)
						srcMAC = &ethernetLayer.SrcMAC
						dstMAC = &ethernetLayer.DstMAC
					default:
						srcMAC = nil
						dstMAC = nil
						fmt.Printf("LINK LAYER: %s\n", linkLayer.LayerType())
					}

					//parse network layer
					networkLayer := packet.NetworkLayer()
					if networkLayer == nil {
						fmt.Printf("No network layer Information, skipping packet\n")
						continue
					}

					switch networkLayer.(type) {
					case *layers.IPv4:
						ipv4Layer, _ := networkLayer.(*layers.IPv4)
						ipProtocol = 4
						srcIP = &ipv4Layer.SrcIP
						dstIP = &ipv4Layer.DstIP
						ipFlags = uint8(ipv4Layer.Flags)
					case *layers.IPv6:
						ipv6Layer, _ := networkLayer.(*layers.IPv6)
						ipProtocol = 6
						srcIP = &ipv6Layer.SrcIP
						dstIP = &ipv6Layer.DstIP
						ipFlags = 0
					default:
						srcIP = nil
						dstIP = nil
						fmt.Printf("NETWORK LAYER: %s\n", networkLayer.LayerType())
					}

					//parse transport layer
					transportLayer := packet.TransportLayer()
					if transportLayer == nil {
						fmt.Printf("No transport layer information, skipping packet\n")
						continue
					}

					switch transportLayer.(type) {
					case *layers.TCP:
						tcpLayer, _ := transportLayer.(*layers.TCP)
						srcPort = uint16(tcpLayer.SrcPort)
						dstPort = uint16(tcpLayer.DstPort)
						sequence = tcpLayer.Seq
						acknowledgement = tcpLayer.Ack
						windowSize = tcpLayer.Window
					case *layers.UDP:
						udpLayer, _ := transportLayer.(*layers.UDP)
						srcPort = uint16(udpLayer.SrcPort)
						dstPort = uint16(udpLayer.DstPort)
						sequence = 0
						acknowledgement = 0
						windowSize = 0
					default:
						srcPort = 0
						dstPort = 0
						sequence = 0
						acknowledgement = 0
						windowSize = 0
						fmt.Printf("TRANSPORT LAYER: %s\n", transportLayer.LayerType())
					}

					//TODO  vlan, TCPFlags
					err = cqlSession.Query(pcapInsertStmt,
							time.Unix(packet.Metadata().Timestamp.Unix() - (packet.Metadata().Timestamp.Unix() % timeBucketSize), 0),
							*captureHostname,
							gocql.UUIDFromTime(packet.Metadata().Timestamp),
							packet.Metadata().Length,
							srcMAC.String(),
							dstMAC.String(),
							ipProtocol,
							srcIP,
							dstIP,
							ipFlags,
							srcPort,
							dstPort,
							"", //TODO tcp_flags text,
							windowSize,
							sequence,
							acknowledgement,
						).Exec()

					if err != nil {
						fmt.Printf("%s\n", err)
					}
				}

				fmt.Printf("duration: %v\n", time.Since(startTime))
			}
		}
	})

	chimpanzee.Run(os.Args)
}
