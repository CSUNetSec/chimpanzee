package converter

import (
	"fmt"
	"net"
	"time"

	cli "github.com/jawher/mow.cli"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func WritePCAPProtobuf(cmd *cli.Cmd) {
	cmd.Spec = "CAPTURE_HOST OUTPUT_FILE FILENAME..."
	captureHostname := cmd.StringArg("CAPTURE_HOST", "", "Host the packets were captured from")
	_ = cmd.StringArg("OUTPUT_FILE", "", "Name of file to write netbrane shared protobufs to")
	filenames := cmd.StringsArg("FILENAME", nil, "Pcap files to process")

	cmd.Action = func() {
		//TODO open outut file
		for _, filename := range *filenames {
			fmt.Printf("working on pcap file '%s'\n", filename)

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

			startTime := time.Now()
			packetCount := 0

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

				fmt.Printf("captureHost:%s\nsrcMAC:%s\ndstMAC:%s\nipProtocol:%d\nsrcIP:%s\ndstIP:%s\nipFlags:%d\nsrcPort:%d\ndstPort:%d\nwindowSize:%d\nseq:%d\nack:%d\n",
					*captureHostname,
					srcMAC.String(),
					dstMAC.String(),
					ipProtocol,
					srcIP,
					dstIP,
					ipFlags,
					srcPort,
					dstPort,
					windowSize,
					sequence,
					acknowledgement)
				//TODO  vlan, TCPFlags
				/*err = cqlSession.Query(pcapInsertStmt,
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
				}*/

				packetCount++
				if packetCount % 2500 == 0 {
					fmt.Printf("packetCount: %d\n", packetCount)
				}
			}

			fmt.Printf("duration: %v packetCount:%d\n", time.Since(startTime), packetCount)
		}
	}
}