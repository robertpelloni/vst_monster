use futures_util::StreamExt;
use reqwest::Client;
use sha2::{Digest, Sha256};
use std::env;
use std::fs;
use std::io;
use std::path::{Path, PathBuf};
use tauri::AppHandle;
use tokio::fs::File;
use tokio::io::AsyncWriteExt;
use zip::ZipArchive;

#[cfg(any(target_os = "windows", target_os = "macos"))]
use tokio::process::Command;

#[tauri::command]
pub async fn install_plugin(
    _app: AppHandle,
    url: String,
    name: String,
    expected_hash: Option<String>,
) -> Result<String, String> {
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
    let mut hasher = Sha256::new();

    while let Some(item) = stream.next().await {
        let chunk = item.map_err(|e| e.to_string())?;
        hasher.update(&chunk);
        file.write_all(&chunk).await.map_err(|e| e.to_string())?;
    }

    // Flush and verify integrity if an expected hash was supplied
    file.flush().await.map_err(|e| e.to_string())?;
    let calculated_hash = hex::encode(hasher.finalize());

    if let Some(target_hash) = expected_hash {
        let trimmed = target_hash.trim().to_lowercase();
        if !trimmed.is_empty() && trimmed != calculated_hash.to_lowercase() {
            // Remove corrupted or compromised file
            let _ = tokio::fs::remove_file(&file_path).await;
            return Err(format!(
                "Integrity check failed: expected sha256 {}, got {}",
                trimmed, calculated_hash
            ));
        }
    }

    println!("Downloaded to {:?}", file_path);

    // If file is a zip archive, extract audio plugin binaries to platform standard paths
    if ext.eq_ignore_ascii_case("zip") {
        extract_zip_archive(&file_path, &safe_name).map_err(|e| format!("Failed to extract ZIP: {}", e))?;
        return Ok(format!("Successfully unpacked and installed {}", name));
    }

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

#[tauri::command]
pub async fn remove_plugin(_app: AppHandle, name: String) -> Result<String, String> {
    let safe_name = name
        .chars()
        .map(|c| if c.is_ascii_alphanumeric() { c } else { '_' })
        .collect::<String>();

    let home = env::var("HOME").or_else(|_| env::var("USERPROFILE")).unwrap_or_else(|_| ".".to_string());
    let home_path = PathBuf::from(home);

    #[cfg(target_os = "linux")]
    let target_dir = home_path.join(".vst3").join(&safe_name);

    #[cfg(target_os = "macos")]
    let target_dir = home_path.join("Library/Audio/Plug-Ins/VST3").join(&safe_name);

    #[cfg(target_os = "windows")]
    let target_dir = home_path.join("Documents/VST3").join(&safe_name);

    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    let target_dir = env::temp_dir().join("vst_monster_plugins").join(&safe_name);

    if target_dir.exists() {
        fs::remove_dir_all(&target_dir).map_err(|e| format!("Failed to remove directory: {}", e))?;
        Ok(format!("Plugin {} removed from {:?}", name, target_dir))
    } else {
        Ok(format!("Plugin files not found at expected directory"))
    }
}

#[derive(serde::Serialize, serde::Deserialize, Clone, Debug)]
pub struct InstalledPluginInfo {
    pub name: String,
    pub version: Option<String>,
    pub format: String,
    pub path: String,
}

#[tauri::command]
pub fn get_default_scan_paths() -> Vec<String> {
    let home = env::var("HOME").or_else(|_| env::var("USERPROFILE")).unwrap_or_else(|_| ".".to_string());
    let home_path = PathBuf::from(home);

    let mut plugin_dirs: Vec<PathBuf> = Vec::new();

    #[cfg(target_os = "linux")]
    {
        plugin_dirs.push(home_path.join(".vst3"));
        plugin_dirs.push(home_path.join(".vst"));
        plugin_dirs.push(home_path.join(".clap"));
        plugin_dirs.push(PathBuf::from("/usr/lib/vst3"));
        plugin_dirs.push(PathBuf::from("/usr/local/lib/vst3"));
    }

    #[cfg(target_os = "macos")]
    {
        plugin_dirs.push(home_path.join("Library/Audio/Plug-Ins/VST3"));
        plugin_dirs.push(home_path.join("Library/Audio/Plug-Ins/Components"));
        plugin_dirs.push(home_path.join("Library/Audio/Plug-Ins/CLAP"));
        plugin_dirs.push(PathBuf::from("/Library/Audio/Plug-Ins/VST3"));
        plugin_dirs.push(PathBuf::from("/Library/Audio/Plug-Ins/Components"));
    }

    #[cfg(target_os = "windows")]
    {
        plugin_dirs.push(home_path.join("Documents/VST3"));
        if let Ok(prog_files) = env::var("CommonProgramFiles") {
            plugin_dirs.push(PathBuf::from(prog_files).join("VST3"));
        }
    }

    plugin_dirs
        .into_iter()
        .map(|p| p.to_string_lossy().to_string())
        .collect()
}

#[tauri::command]
pub async fn scan_installed_plugins(_app: AppHandle, extra_paths: Option<Vec<String>>) -> Result<Vec<InstalledPluginInfo>, String> {
    let home = env::var("HOME").or_else(|_| env::var("USERPROFILE")).unwrap_or_else(|_| ".".to_string());
    let home_path = PathBuf::from(home);

    let mut plugin_dirs: Vec<PathBuf> = Vec::new();

    #[cfg(target_os = "linux")]
    {
        plugin_dirs.push(home_path.join(".vst3"));
        plugin_dirs.push(home_path.join(".vst"));
        plugin_dirs.push(home_path.join(".clap"));
        plugin_dirs.push(PathBuf::from("/usr/lib/vst3"));
        plugin_dirs.push(PathBuf::from("/usr/local/lib/vst3"));
    }

    #[cfg(target_os = "macos")]
    {
        plugin_dirs.push(home_path.join("Library/Audio/Plug-Ins/VST3"));
        plugin_dirs.push(home_path.join("Library/Audio/Plug-Ins/Components"));
        plugin_dirs.push(home_path.join("Library/Audio/Plug-Ins/CLAP"));
        plugin_dirs.push(PathBuf::from("/Library/Audio/Plug-Ins/VST3"));
        plugin_dirs.push(PathBuf::from("/Library/Audio/Plug-Ins/Components"));
    }

    #[cfg(target_os = "windows")]
    {
        plugin_dirs.push(home_path.join("Documents/VST3"));
        if let Ok(prog_files) = env::var("CommonProgramFiles") {
            plugin_dirs.push(PathBuf::from(prog_files).join("VST3"));
        }
    }

    if let Some(extras) = extra_paths {
        for ep in extras {
            let path_buf = PathBuf::from(ep.trim());
            if !path_buf.as_os_str().is_empty() && !plugin_dirs.contains(&path_buf) {
                plugin_dirs.push(path_buf);
            }
        }
    }

    let mut discovered: Vec<InstalledPluginInfo> = Vec::new();

    for dir in plugin_dirs {
        if dir.exists() && dir.is_dir() {
            if let Ok(entries) = fs::read_dir(&dir) {
                for entry in entries.flatten() {
                    let path = entry.path();
                    let file_name = entry.file_name().to_string_lossy().to_string();

                    let format = if file_name.ends_with(".vst3") || path.extension().map_or(false, |ext| ext == "vst3") {
                        "vst3".to_string()
                    } else if file_name.ends_with(".clap") {
                        "clap".to_string()
                    } else if file_name.ends_with(".component") {
                        "au".to_string()
                    } else if file_name.ends_with(".so") || file_name.ends_with(".dll") {
                        "vst2".to_string()
                    } else if path.is_dir() {
                        "vst3".to_string()
                    } else {
                        continue;
                    };

                    let clean_name = file_name
                        .trim_end_matches(".vst3")
                        .trim_end_matches(".so")
                        .trim_end_matches(".component")
                        .trim_end_matches(".clap")
                        .trim_end_matches(".dll")
                        .to_string();

                    if clean_name.is_empty() {
                        continue;
                    }

                    // Attempt to detect version from bundle if directory or metadata file exists
                    let mut detected_version: Option<String> = None;
                    if path.is_dir() {
                        // Check for common version marker files or info files
                        let version_file = path.join("VERSION");
                        let version_txt = path.join("version.txt");
                        let pkg_json = path.join("package.json");
                        if version_file.exists() {
                            if let Ok(v) = fs::read_to_string(&version_file) {
                                detected_version = Some(v.trim().to_string());
                            }
                        } else if version_txt.exists() {
                            if let Ok(v) = fs::read_to_string(&version_txt) {
                                detected_version = Some(v.trim().to_string());
                            }
                        } else if pkg_json.exists() {
                            if let Ok(v) = fs::read_to_string(&pkg_json) {
                                if let Ok(parsed) = serde_json::from_str::<serde_json::Value>(&v) {
                                    if let Some(ver) = parsed.get("version").and_then(|val| val.as_str()) {
                                        detected_version = Some(ver.to_string());
                                    }
                                }
                            }
                        }
                    }

                    // Avoid duplicate entries
                    if !discovered.iter().any(|item| item.name.eq_ignore_ascii_case(&clean_name)) {
                        discovered.push(InstalledPluginInfo {
                            name: clean_name,
                            version: detected_version,
                            format,
                            path: path.to_string_lossy().to_string(),
                        });
                    }
                }
            }
        }
    }

    Ok(discovered)
}

#[tauri::command]
pub async fn verify_installed_plugin(_app: AppHandle, path: String) -> Result<bool, String> {
    let p = PathBuf::from(&path);
    if !p.exists() {
        return Ok(false);
    }

    if p.is_file() {
        if let Ok(meta) = fs::metadata(&p) {
            return Ok(meta.len() > 0);
        }
        return Ok(false);
    }

    if p.is_dir() {
        // Inspect if directory contains valid files or nested audio bundle structure
        if let Ok(entries) = fs::read_dir(&p) {
            let mut count = 0;
            for entry in entries.flatten() {
                count += 1;
                let subpath = entry.path();
                if subpath.is_file() {
                    return Ok(true);
                }
                // Check standard VST3 bundle structure: Contents/
                if subpath.is_dir() && entry.file_name().to_string_lossy().eq_ignore_ascii_case("contents") {
                    return Ok(true);
                }
            }
            return Ok(count > 0);
        }
    }

    Ok(false)
}

#[tauri::command]
pub async fn open_plugin_directory(_app: AppHandle, name: Option<String>) -> Result<String, String> {
    let home = env::var("HOME").or_else(|_| env::var("USERPROFILE")).unwrap_or_else(|_| ".".to_string());
    let home_path = PathBuf::from(home);

    #[cfg(target_os = "linux")]
    let base_dir = home_path.join(".vst3");

    #[cfg(target_os = "macos")]
    let base_dir = home_path.join("Library/Audio/Plug-Ins/VST3");

    #[cfg(target_os = "windows")]
    let base_dir = home_path.join("Documents/VST3");

    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    let base_dir = env::temp_dir().join("vst_monster_plugins");

    let target = if let Some(n) = name {
        let safe_name = n
            .chars()
            .map(|c| if c.is_ascii_alphanumeric() { c } else { '_' })
            .collect::<String>();
        let specific_dir = base_dir.join(&safe_name);
        if specific_dir.exists() {
            specific_dir
        } else {
            base_dir
        }
    } else {
        base_dir
    };

    if !target.exists() {
        let _ = fs::create_dir_all(&target);
    }

    #[cfg(target_os = "linux")]
    {
        tokio::process::Command::new("xdg-open")
            .arg(target.to_str().unwrap_or("."))
            .spawn()
            .map_err(|e| e.to_string())?;
    }

    #[cfg(target_os = "macos")]
    {
        tokio::process::Command::new("open")
            .arg(target.to_str().unwrap_or("."))
            .spawn()
            .map_err(|e| e.to_string())?;
    }

    #[cfg(target_os = "windows")]
    {
        tokio::process::Command::new("explorer")
            .arg(target.to_str().unwrap_or("."))
            .spawn()
            .map_err(|e| e.to_string())?;
    }

    Ok(format!("Opened {:?}", target))
}

fn extract_zip_archive(zip_path: &Path, plugin_name: &str) -> Result<(), io::Error> {
    let file = fs::File::open(zip_path)?;
    let mut archive = ZipArchive::new(file)
        .map_err(|e| io::Error::new(io::ErrorKind::Other, e.to_string()))?;

    // Determine target plugin directory based on current OS
    let home = env::var("HOME").or_else(|_| env::var("USERPROFILE")).unwrap_or_else(|_| ".".to_string());
    let home_path = PathBuf::from(home);

    #[cfg(target_os = "linux")]
    let target_dir = home_path.join(".vst3").join(plugin_name);

    #[cfg(target_os = "macos")]
    let target_dir = home_path.join("Library/Audio/Plug-Ins/VST3").join(plugin_name);

    #[cfg(target_os = "windows")]
    let target_dir = home_path.join("Documents/VST3").join(plugin_name);

    #[cfg(not(any(target_os = "linux", target_os = "macos", target_os = "windows")))]
    let target_dir = env::temp_dir().join("vst_monster_plugins").join(plugin_name);

    fs::create_dir_all(&target_dir)?;

    for i in 0..archive.len() {
        let mut file = archive.by_index(i)
            .map_err(|e| io::Error::new(io::ErrorKind::Other, e.to_string()))?;
        let outpath = match file.enclosed_name() {
            Some(path) => target_dir.join(path),
            None => continue,
        };

        if file.name().ends_with('/') {
            fs::create_dir_all(&outpath)?;
        } else {
            if let Some(p) = outpath.parent() {
                if !p.exists() {
                    fs::create_dir_all(p)?;
                }
            }
            let mut outfile = fs::File::create(&outpath)?;
            io::copy(&mut file, &mut outfile)?;
        }
    }

    Ok(())
}
