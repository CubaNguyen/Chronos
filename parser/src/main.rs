use std::path::Path;
use std::process::ExitCode;

use clap::{Parser, Subcommand};
use serde::Serialize;
use walkdir::{DirEntry, WalkDir};

use chronos_parser::ast::Ast;
use chronos_parser::detector::detect;
use chronos_parser::lang::factory::create_parser;

#[derive(Parser)]
#[command(name = "chronos-parser")]
#[command(version = "0.1.0")]
#[command(about = "Parse source code into normalized AST")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Parse a file or directory
    Parse {
        /// Path to file or directory
        path: String,

        /// Output compact JSON (single line) instead of pretty
        #[arg(long)]
        compact: bool,
    },
}

fn main() -> ExitCode {
    let cli = Cli::parse();

    let result = match cli.command {
        Commands::Parse { path, compact } => cmd_parse(&path, compact),
    };

    match result {
        Ok(()) => ExitCode::SUCCESS,
        Err(e) => {
            eprintln!("error: {}", e);
            ExitCode::FAILURE
        }
    }
}

fn cmd_parse(path: &str, compact: bool) -> Result<(), Box<dyn std::error::Error>> {
    let p = Path::new(path);

    if !p.exists() {
        return Err(format!("path does not exist: {}", path).into());
    }

    if p.is_file() {
        let ast = parse_one(p)?;
        print_json(&ast, compact)?;
    } else if p.is_dir() {
        let asts = parse_dir(p)?;
        print_json(&asts, compact)?;
    } else {
        return Err(format!("path is neither file nor directory: {}", path).into());
    }

    Ok(())
}

/// Parse 1 file. Trả về lỗi nếu file không hỗ trợ hoặc parse fail.
fn parse_one(path: &Path) -> Result<Ast, Box<dyn std::error::Error>> {
    let path_str = path.to_string_lossy().to_string();
    let source = std::fs::read_to_string(path)?;
    let lang = detect(&path_str)?;
    let parser = create_parser(lang)?;
    let ast = parser.parse(&source, &path_str)?;
    Ok(ast)
}

/// Walk directory, parse mọi file có ngôn ngữ được hỗ trợ.
/// File không hỗ trợ → skip silently (đếm vào `unsupported`).
/// File hỗ trợ nhưng parse fail → log warning ra stderr (đếm vào `failed`).
fn parse_dir(dir: &Path) -> Result<Vec<Ast>, Box<dyn std::error::Error>> {
    let mut results: Vec<Ast> = Vec::new();
    let mut unsupported = 0usize;
    let mut failed = 0usize;

    let walker = WalkDir::new(dir)
        .follow_links(false)
        .into_iter()
        .filter_entry(|e| !should_skip(e));

    for entry in walker {
        let entry = entry?;
        if !entry.file_type().is_file() {
            continue;
        }

        let path = entry.path();
        let path_str = path.to_string_lossy().to_string();

        // Check detect trước — file không hỗ trợ thì skip
        if detect(&path_str).is_err() {
            unsupported += 1;
            continue;
        }

        match parse_one(path) {
            Ok(ast) => results.push(ast),
            Err(e) => {
                eprintln!("warning: failed to parse {}: {}", path_str, e);
                failed += 1;
            }
        }
    }

    eprintln!(
        "parsed {} files (skipped: {} unsupported, {} failed)",
        results.len(),
        unsupported,
        failed
    );

    Ok(results)
}

/// Skip hidden files/dirs và các thư mục rác.
fn should_skip(entry: &DirEntry) -> bool {
    let name = entry.file_name().to_string_lossy();

    // Bỏ qua mọi thứ bắt đầu bằng dấu chấm (.git, .env, .cache...)
    if name.starts_with('.') {
        return true;
    }

    // Bỏ qua thư mục build/dependency
    if entry.file_type().is_dir() {
        matches!(
            name.as_ref(),
            "target" | "node_modules" | "vendor" | "dist" | "build"
        )
    } else {
        false
    }
}

fn print_json<T: Serialize>(value: &T, compact: bool) -> Result<(), Box<dyn std::error::Error>> {
    let s = if compact {
        serde_json::to_string(value)?
    } else {
        serde_json::to_string_pretty(value)?
    };
    println!("{}", s);
    Ok(())
}