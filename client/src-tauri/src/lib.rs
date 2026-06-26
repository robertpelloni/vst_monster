use std::fs::File;
use std::io::Write;
use std::path::PathBuf;
use tauri::AppHandle;

#[tauri::command]
async fn install_plugin(url: String, name: String, _app: AppHandle) -> Result<String, String> {
    println!("Starting installation for {} from {}", name, url);

    // Create a temporary file path
    let mut temp_dir = std::env::temp_dir();
    temp_dir.push(format!("{}.vst_download", name));

    // Perform the download
    match download_file(&url, &temp_dir).await {
        Ok(_) => Ok(format!("Successfully downloaded {} to {:?}", name, temp_dir)),
        Err(e) => Err(format!("Failed to download plugin: {}", e))
    }
}

async fn download_file(url: &str, path: &PathBuf) -> Result<(), Box<dyn std::error::Error>> {
    let response = reqwest::get(url).await?;

    if !response.status().is_success() {
        return Err(format!("Server returned error: {}", response.status()).into());
    }

    let bytes = response.bytes().await?;
    let mut file = File::create(path)?;
    file.write_all(&bytes)?;

    Ok(())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .invoke_handler(tauri::generate_handler![install_plugin])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
