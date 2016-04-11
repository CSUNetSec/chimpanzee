CREATE KEYSPACE netbrane_pcap_core WITH replication = {'class': 'NetworkTopologyStrategy', 'csu_dc': 2};

CREATE TABLE netbrane_pcap_core.packets_by_time (
	time_bucket timestamp,
	capture_host text,
	timestamp timeuuid,
	packet_size int,
	source_mac text,
	destination_mac text,
	ip_protocol int,
	source_ip inet,
	destination_ip inet,
	ip_flags int,
	source_port int,
	destination_port int,
	tcp_flags text,
	tcp_window_size int,
	tcp_sequence bigint,
	tcp_acknowledgement bigint,
	PRIMARY KEY ((time_bucket, capture_host), timestamp)
) WITH CLUSTERING ORDER BY (timestamp DESC);
