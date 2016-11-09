use std::fs::File;
use std::io::{BufRead, BufReader};
use std::iter::Iterator;
use std::net::IpAddr;
use std::str::FromStr;

use pb::common::IPAddressWrapper;
use pb::netbrane::{CaptureRecordUnion, CaptureRecordUnion_RecordType, DNSRecord};
use pb::proddle::{Probe, ProbeResult};

use protobuf::{CodedInputStream, Message, ProtobufError, RepeatedField};

/*
 * DNSReader
 */
pub struct DNSReader<'a> {
    buf_reader: BufReader<&'a mut File>,
    line: String,
}

impl<'a> DNSReader<'a> {
    pub fn new(file: &'a mut File) -> DNSReader<'a> {
        DNSReader {
            buf_reader: BufReader::new(file),
            line: String::new(),
        }
    }
}

impl<'a> Iterator for DNSReader<'a> {
    type Item = CaptureRecordUnion;

    fn next(&mut self) -> Option<CaptureRecordUnion> {
        self.line.clear();
        if let Err(_) = self.buf_reader.read_line(&mut self.line) {
            return None;
        }

        //parse fields
        let vec: Vec<&str> = self.line.trim().split(":").collect();

        //create dns record
        let mut dns_record = DNSRecord::new();
        dns_record.set_query(vec[3].to_owned());
        
        //set requesting host
        if let Ok(ip_addr) = IpAddr::from_str(vec[1]) {
            let mut ip_address_wrapper = IPAddressWrapper::new();
            match ip_addr {
                IpAddr::V4(ipv4_addr) => ip_address_wrapper.set_ipv4(ipv4_addr.octets().to_vec()),
                IpAddr::V6(ipv6_addr) => ip_address_wrapper.set_ipv6(ipv6_addr.octets().to_vec()),
            }

            dns_record.set_requesting_host(ip_address_wrapper);
        }

        //set dns server
        if let Ok(ip_addr) = IpAddr::from_str(vec[2]) {
            let mut ip_address_wrapper = IPAddressWrapper::new();
            match ip_addr {
                IpAddr::V4(ipv4_addr) => ip_address_wrapper.set_ipv4(ipv4_addr.octets().to_vec()),
                IpAddr::V6(ipv6_addr) => ip_address_wrapper.set_ipv6(ipv6_addr.octets().to_vec()),
            }

            dns_record.set_dns_server(ip_address_wrapper);
        }

        //parse ip addresses
        let mut ips = RepeatedField::new();
        let mut cname = RepeatedField::new();
        for ip_string in vec[vec.len() - 1].split(" ") {
            match IpAddr::from_str(ip_string) {
                Ok(ip_addr) => {
                    let mut ip_address_wrapper = IPAddressWrapper::new();
                    match ip_addr {
                        IpAddr::V4(ipv4_addr) => ip_address_wrapper.set_ipv4(ipv4_addr.octets().to_vec()),
                        IpAddr::V6(ipv6_addr) => ip_address_wrapper.set_ipv6(ipv6_addr.octets().to_vec()),
                    }

                    ips.push(ip_address_wrapper);
                },
                Err(_) => {
                    if ip_string != "" {
                        cname.push(ip_string.to_owned());
                    }
                },
            }
        }

        dns_record.set_ips(ips);
        if cname.len() != 0 {
            dns_record.set_cname(cname);
        }

        //parse timestamp
        let timestamp_seconds = match i64::from_str(vec[0]) {
            Ok(timestamp_seconds) => timestamp_seconds,
            Err(e) => panic!("{}", e),
        };

        let mut capture_record_union = CaptureRecordUnion::new();
        capture_record_union.set_timestamp_seconds(timestamp_seconds);
        capture_record_union.set_record_type(CaptureRecordUnion_RecordType::DNS_RECORD);
        capture_record_union.set_dns_record(dns_record);
        Some(capture_record_union)
    }
}

/*
 * ProbeReader
 */
pub struct ProbeReader<'a> {
    coded_input_stream: CodedInputStream<'a>,
}

impl<'a> ProbeReader<'a> {
    pub fn new(file: &'a mut File) -> ProbeReader<'a> {
        ProbeReader {
            coded_input_stream: CodedInputStream::new(file),
        }
    }
}

impl<'a> Iterator for ProbeReader<'a> {
    type Item = CaptureRecordUnion;

    fn next(&mut self) -> Option<CaptureRecordUnion> {
        //check for end of file
        if self.coded_input_stream.eof().unwrap() {
            return None;
        }

        //parse probe
        let mut probe = Probe::new();
        let _ = read_protobuf(&mut self.coded_input_stream, &mut probe);

        let mut capture_record_union = CaptureRecordUnion::new();
        capture_record_union.set_record_type(CaptureRecordUnion_RecordType::PROBE_RECORD);
        capture_record_union.set_probe_record(probe);
        Some(capture_record_union)
    }
}

/*
 * ProbeResultReader
 */
pub struct ProbeResultReader<'a> {
    coded_input_stream: CodedInputStream<'a>,
}

impl<'a> ProbeResultReader<'a> {
    pub fn new(file: &'a mut File) -> ProbeResultReader<'a> {
        ProbeResultReader {
            coded_input_stream: CodedInputStream::new(file),
        }
    }
}

impl<'a> Iterator for ProbeResultReader<'a> {
    type Item = CaptureRecordUnion;

    fn next(&mut self) -> Option<CaptureRecordUnion> {
        //check for end of file
        if self.coded_input_stream.eof().unwrap() {
            return None;
        }

        //parse probe result
        let mut probe_result = ProbeResult::new();
        let _ = read_protobuf(&mut self.coded_input_stream, &mut probe_result);

        let mut capture_record_union = CaptureRecordUnion::new();
        capture_record_union.set_record_type(CaptureRecordUnion_RecordType::PROBE_RESULT_RECORD);
        capture_record_union.set_probe_result_record(probe_result);
        Some(capture_record_union)
    }
}

/*
 * Misc Helpers
 */
fn read_protobuf(coded_input_stream: &mut CodedInputStream, message: &mut Message) -> Result<(), ProtobufError> {
    //read length
    let length = try!(coded_input_stream.read_uint32());

    //read bytes for messages
    let mut bytes = Vec::new();
    for _ in 0..length {
        let byte = try!(coded_input_stream.read_raw_byte());
        bytes.push(byte);
    }

    //parse message
    let mut message_input_stream = CodedInputStream::from_bytes(&bytes);
    let _ = message.merge_from(&mut message_input_stream);
    Ok(())
}
