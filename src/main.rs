extern crate docopt;
extern crate protobuf;
extern crate rustc_serialize;

pub mod pb;

use docopt::Docopt;

const USAGE: &'static str = "
Conversion tool to netbrane common format.

Usage:
    chimpanzee (-h | --help)

Options:
    -h --help               Display this help message.
";

#[derive(Debug, RustcDecodable)]
struct Args {
}

fn main() {
    let args: Args = Docopt::new(USAGE)
                        .and_then(|d| d.decode())
                        .unwrap_or_else(|e| e.exit());

    println!("don't mess with the chimpanzee!");
}
