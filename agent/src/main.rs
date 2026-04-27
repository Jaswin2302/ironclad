use sysinfo::System;
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use serde::Serialize;
use tokio::net::UnixListener;
use tokio::io::AsyncWriteExt;

#[derive(Serialize)]
struct Metrics {
    timestamp: u64,
    cpu_percent: f64,
    mem_percent: f64,
    mem_used_mb: u64,
    mem_total_mb: u64,
}

fn timestamp() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_secs()
}

fn collect_metrics(sys: &mut System) -> Metrics {
    sys.refresh_all();
    let total_mem = sys.total_memory();
    let used_mem = sys.used_memory();
    Metrics {
        timestamp: timestamp(),
        cpu_percent: sys.cpus().iter().map(|c| c.cpu_usage() as f64).sum::<f64>() / sys.cpus().len() as f64,
        mem_percent: (used_mem as f64 / total_mem as f64) * 100.0,
        mem_used_mb: used_mem / 1024 / 1024,
        mem_total_mb: total_mem / 1024 / 1024,
    }
}

#[tokio::main]
async fn main() {
    let socket_path = "/tmp/ironclad.sock";

    // Remove old socket file if it exists
    let _ = std::fs::remove_file(socket_path);

    let listener = UnixListener::bind(socket_path).expect("Failed to bind socket");
    println!("[ironclad-agent] Listening on {}", socket_path);

    let mut sys = System::new_all();

    loop {
        // Wait for a controller to connect
        let (mut stream, _) = listener.accept().await.expect("Failed to accept connection");
        println!("[ironclad-agent] Controller connected");

        // Keep sending metrics to this connected controller
        loop {
            let metrics = collect_metrics(&mut sys);
            let mut json = serde_json::to_string(&metrics).expect("Failed to serialize");
            json.push('\n'); // newline delimiter so controller knows where each message ends

            if stream.write_all(json.as_bytes()).await.is_err() {
                println!("[ironclad-agent] Controller disconnected");
                break;
            }

            tokio::time::sleep(Duration::from_secs(2)).await;
        }
    }
}