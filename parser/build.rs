use std::env;
use std::path::PathBuf;

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let out_dir = PathBuf::from(env::var("OUT_DIR")?);
    let descriptor_path = out_dir.join("chronos_parser_descriptor.bin");

    tonic_prost_build::configure()
        .build_server(true)
        .build_client(false)
        .file_descriptor_set_path(descriptor_path)   // ← ĐỔI: dùng absolute path
        .compile_protos(
            &["../proto/parser.proto"],
            &["../proto"],
        )?;

    println!("cargo:rerun-if-changed=../proto/parser.proto");

    Ok(())
}