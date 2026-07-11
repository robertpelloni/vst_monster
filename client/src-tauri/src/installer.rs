use futures_util::StreamExt;
use reqwest::Client;
use std::env;
use std::path::PathBuf;
use tauri::AppHandle;
use tokio::fs::File;
use tokio::io::AsyncWriteExt;

#[cfg(any(target_os = "windows", target_os = "macos"))]
use tokio::process::Command;

#[tauri::command]
pub async fn install_plugin(_app: AppHandle, url: String, name: String) -> Result<String, String> {
    let client = Client::new();
    let res = client.get(&url).send().await.map_err(|e| e.to_string())?;

    if !res.status().is_success() {
        return Err(format!("Failed to download plugin: HTTP {}", res.status()));
    }

    let tmp_dir = env::temp_dir();

    // Safely extract extension, defaulting to zip and ensuring it only contains alphanumeric chars
    let raw_ext = url
        .split('?')
        .next()
        .unwrap_or(&url)
        .split('.')
        .last()
        .unwrap_or("zip");
    let safe_ext: String = raw_ext
        .chars()
        .filter(|c| c.is_ascii_alphanumeric())
        .collect();
    let ext = if safe_ext.is_empty() {
        "zip".to_string()
    } else {
        safe_ext
    };

    // Sanitize name to prevent path traversal
    let safe_name = name
        .chars()
        .map(|c| if c.is_ascii_alphanumeric() { c } else { '_' })
        .collect::<String>();

    let file_path: PathBuf = tmp_dir.join(format!("{}.{}", safe_name, ext));

    let mut file = File::create(&file_path).await.map_err(|e| e.to_string())?;
    let mut stream = res.bytes_stream();

    while let Some(item) = stream.next().await {
        let chunk = item.map_err(|e| e.to_string())?;
        file.write_all(&chunk).await.map_err(|e| e.to_string())?;
    }

    println!("Downloaded to {:?}", file_path);

    // Platform-specific installation logic using tokio::process::Command to avoid blocking the thread
    #[cfg(target_os = "windows")]
    {
        if ext.eq_ignore_ascii_case("msi") {
            let status = Command::new("msiexec")
                .args(&[
                    "/i",
                    file_path.to_str().unwrap(),
                    "/quiet",
                    "/qn",
                    "/norestart",
                ])
                .status()
                .await
                .map_err(|e| e.to_string())?;
            if !status.success() {
                return Err(format!("MSI installation failed with status: {}", status));
            }
        } else if ext.eq_ignore_ascii_case("exe") {
            let status = Command::new(file_path.to_str().unwrap())
                .args(&["/S", "/silent", "/quiet"])
                .status()
                .await
                .map_err(|e| e.to_string())?;
            if !status.success() {
                return Err(format!("EXE installation failed with status: {}", status));
            }
        }
    }

    #[cfg(target_os = "macos")]
    {
        if ext.eq_ignore_ascii_case("pkg") {
            let status = Command::new("installer")
                .args(&["-pkg", file_path.to_str().unwrap(), "-target", "/"])
                .status()
                .await
                .map_err(|e| e.to_string())?;
            if !status.success() {
                return Err(format!("PKG installation failed with status: {}", status));
            }
        } else if ext.eq_ignore_ascii_case("dmg") {
            // DMG Extraction logic (Mount, Copy, Unmount)
            let attach_status = Command::new("hdiutil")
                .args(&["attach", "-plist", file_path.to_str().unwrap()])
                .output()
                .await
                .map_err(|e| e.to_string())?;

            if !attach_status.status.success() {
                return Err(format!("Failed to mount DMG"));
            }

            // In a real implementation, you'd parse the plist to find the mount point,
            // copy the .vst3/.component files to /Library/Audio/Plug-Ins/ or ~/Library/Audio/Plug-Ins/
            // and then detach the volume using `hdiutil detach`.
            // This is a simplified stub.
            return Ok(format!(
                "Downloaded DMG. Automatic DMG extraction is a stub."
            ));
        }
    }

    Ok(format!("Successfully installed {}", name))
}
