extern crate docopt;
extern crate protobuf;
extern crate rustc_serialize;

use docopt::Docopt;
use protobuf::{CodedOutputStream, Message};

pub mod pb;
pub mod reader;

use pb::netbrane::CaptureRecordUnion;
use reader::{ProbeReader, ProbeResultReader};

use std::fs::File;
use std::iter::Iterator;

const USAGE: &'static str = "
Conversion tool to netbrane common format.

Usage:
    chimpanzee convert (--probe | --probe-result) <infile> [--outfile=<outfile>]
    chimpanzee (-h | --help)

Options:
    --outfile=<outfile>     Output file.
    --probe                 Input file type is probe.
    --probe_result          Input file type is probe result.
    -h --help               Display this help message.
";

#[derive(Debug, RustcDecodable)]
struct Args {
    cmd_convert: bool,
    arg_infile: String,
    flag_outfile: Option<String>,
    flag_probe: bool,
    flag_probe_result: bool,
}

fn main() {
    let args: Args = Docopt::new(USAGE)
                        .and_then(|d| d.decode())
                        .unwrap_or_else(|e| e.exit());

    if args.cmd_convert {
        //open file reader
        let mut file = match File::open(&args.arg_infile) {
            Ok(file) => file,
            Err(e) => panic!("{}", e),
        };

        let iter: Box<Iterator<Item=CaptureRecordUnion>> = if args.flag_probe {
            Box::new(ProbeReader::new(&mut file))
        } else if args.flag_probe_result {
            Box::new(ProbeResultReader::new(&mut file))
        } else {
            panic!("unknown input file type");
        };

        //open file writer
        let mut out_file = match File::create(args.flag_outfile.unwrap_or(format!("{}.unb", args.arg_infile))) {
            Ok(out_file) => out_file,
            Err(e) => panic!("{}", e),
        };

        let mut coded_output_stream = CodedOutputStream::new(&mut out_file);

        //convert records
        for capture_record_union in iter {
            if let Err(e) = coded_output_stream.write_uint32_no_tag(capture_record_union.compute_size()) {
                panic!("{}", e);
            }

            if let Err(e) = capture_record_union.write_to_with_cached_sizes(&mut coded_output_stream) {
                panic!("{}", e);
            }
        }

        if let Err(e) = coded_output_stream.flush() {
            panic!("{}", e);
        }
    }
}
