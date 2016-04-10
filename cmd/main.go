package main

import (
	"fmt"
	"net"
	"os"

	cli "github.com/jawher/mow.cli"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// ./chimpanzee -uusername -ppassword -h1.2.3.4 -h1.2.3.5 pcap 1.pcap 2.pcap 3.pcap

func main() {
	chimpanzee := cli.App("chimpanzee", "Write data to netbrane defined cassandra tables")
	chimpanzee.Version("v version", "0.0.1")

	cassandraUsername := chimpanzee.StringOpt("u username", "", "Cassandra username")
	cassandraPassword := chimpanzee.StringOpt("p password", "", "Cassandra password")
	cassandraHosts := chimpanzee.StringsOpt("h host", nil, "Cassandra host IPs")

	//process pcap files
	chimpanzee.Command("pcap", "Write a pcap file", func(cmd *cli.Cmd) {
		cmd.Spec = "FILENAME..."
		filenames := cmd.StringsArg("FILENAME", nil, "pcap files to write")

		cmd.Action = func() {
			for _, filename := range *filenames {
				fmt.Printf("writing pcap file '%s' to '%v' username:%s password:%s\n", filename, *cassandraHosts, *cassandraUsername, *cassandraPassword)

				//open pcap file
				handle, err := pcap.OpenOffline(filename)
				if err != nil {
					fmt.Printf("%s\n", err)
					continue
				}

				defer handle.Close()

				//loop through pcap packets
				var (
					packetLength uint16
					srcMAC, dstMAC *net.HardwareAddr
					ipProtocol int
					srcIP, dstIP *net.IP
					ipFlags uint8
					srcPort, dstPort uint16
					sequence, acknowledgement uint32
					windowSize uint16
				)
				packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
				for packet := range packetSource.Packets() {
					//parse link layer
					linkLayer := packet.LinkLayer()
					switch linkLayer.(type) {
					case *layers.Ethernet:
						ethernetLayer, _ := linkLayer.(*layers.Ethernet)
						packetLength = ethernetLayer.Length
						srcMAC = &ethernetLayer.SrcMAC
						dstMAC = &ethernetLayer.DstMAC
					default:
						srcMAC = nil
						dstMAC = nil
						fmt.Printf("LINK LAYER: %s", linkLayer.LayerType())
					}

					//parse network layer
					networkLayer := packet.NetworkLayer()
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
						fmt.Printf("NETWORK LAYER: %s", networkLayer.LayerType())
					}

					//parse transport layer
					transportLayer := packet.TransportLayer()
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
						fmt.Printf("TRANSPORT LAYER: %s", transportLayer.LayerType())
					}

					//TODO  vlan, TCPFlags
					fmt.Printf("Length:%d SrcMac:%s DestMac:%s\n", packetLength, srcMAC, dstMAC)
					fmt.Printf("IPProto:%d SrcIP:%s DestIP:%s Flags:%d\n", ipProtocol, srcIP, dstIP, ipFlags)
					fmt.Printf("SrcPort:%d DstPort:%d seq:%d, ack:%d, window:%d\n", srcPort, dstPort, sequence, acknowledgement, windowSize)
					fmt.Println()
				}
			}
		}
	})

	chimpanzee.Run(os.Args)
}
