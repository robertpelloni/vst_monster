#[allow(unused_imports)]
use std::process::Command;
use std::path::Path;

pub fn install(file_path: &str) -> Result<String, String> {
    let path = Path::new(file_path);

    // Security Nitpick: Ensure we are only installing from within a designated scope, e.g. Downloads directory
    // For this prototype, we'll ensure it resolves to a valid absolute path inside the home directory
    let canonical = path.canonicalize().map_err(|e| format!("Invalid path: {}", e))?;
    let home = std::env::var("HOME").or_else(|_| std::env::var("USERPROFILE")).unwrap_or_default();
    if !home.is_empty() && !canonical.to_string_lossy().starts_with(&home) {
        return Err("Security Error: Target file is outside the allowed user directory scope.".into());
    }

    if !path.exists() {
        return Err(format!("File does not exist: {}", file_path));
    }

    let extension = path.extension().and_then(|e| e.to_str()).unwrap_or("").to_lowercase();

    match extension.as_str() {
        "msi" => install_msi(file_path),
        "exe" => install_exe(file_path),
        "pkg" => install_pkg(file_path),
        "dmg" => install_dmg(file_path),
        _ => Err(format!("Unsupported installer format: .{}", extension)),
    }
}

#[cfg(target_os = "windows")]
fn install_msi(file_path: &str) -> Result<String, String> {
    let output = Command::new("msiexec")
        .args(&["/i", file_path, "/qn"])
        .output()
        .map_err(|e| format!("Failed to execute msiexec: {}", e))?;

    if output.status.success() {
        Ok(format!("Successfully installed MSI: {}", file_path))
    } else {
        Err(format!("MSI installation failed: {:?}", String::from_utf8_lossy(&output.stderr)))
    }
}

#[cfg(not(target_os = "windows"))]
fn install_msi(_file_path: &str) -> Result<String, String> {
    Err("MSI installers are only supported on Windows".into())
}

#[cfg(target_os = "windows")]
fn install_exe(file_path: &str) -> Result<String, String> {
    // Attempt common silent flags
    let output = Command::new(file_path)
        .arg("/S") // common for NSIS
        .output()
        .map_err(|e| format!("Failed to execute EXE: {}", e))?;

    if output.status.success() {
        Ok(format!("Successfully installed EXE: {}", file_path))
    } else {
        Err(format!("EXE installation failed: {:?}", String::from_utf8_lossy(&output.stderr)))
    }
}

#[cfg(not(target_os = "windows"))]
fn install_exe(_file_path: &str) -> Result<String, String> {
    Err("EXE installers are only supported on Windows".into())
}

#[cfg(target_os = "macos")]
fn install_pkg(file_path: &str) -> Result<String, String> {
    let output = Command::new("installer")
        .args(&["-pkg", file_path, "-target", "/"])
        .output()
        .map_err(|e| format!("Failed to execute installer: {}", e))?;

    if output.status.success() {
        Ok(format!("Successfully installed PKG: {}", file_path))
    } else {
        Err(format!("PKG installation failed: {:?}", String::from_utf8_lossy(&output.stderr)))
    }
}

#[cfg(not(target_os = "macos"))]
fn install_pkg(_file_path: &str) -> Result<String, String> {
    Err("PKG installers are only supported on macOS".into())
}

#[cfg(target_os = "macos")]
fn install_dmg(file_path: &str) -> Result<String, String> {
    // Mount the DMG
    let attach_output = Command::new("hdiutil")
        .args(&["attach", "-plist", file_path])
        .output()
        .map_err(|e| format!("Failed to attach DMG: {}", e))?;

    if !attach_output.status.success() {
        return Err(format!("DMG attach failed: {:?}", String::from_utf8_lossy(&attach_output.stderr)));
    }

    let stdout = String::from_utf8_lossy(&attach_output.stdout);
    // Extract mount point naively from plist output or regular output
    // Simple fallback string parsing if we don't want to bring in a plist parser just for this
    let mut mount_point = "";
    for line in stdout.lines() {
        if line.contains("/Volumes/") {
            if let Some(idx) = line.find("/Volumes/") {
                let end_idx = line[idx..].find('<').map(|i| i + idx).unwrap_or(line.len());
                mount_point = &line[idx..end_idx];
                break;
            }
        }
    }

    if mount_point.is_empty() {
        // Fallback rudimentary search
        mount_point = stdout.lines().last().unwrap_or("").split('\t').last().unwrap_or("").trim();
    }

    let mut copied_files = Vec::new();

    if !mount_point.is_empty() && Path::new(mount_point).exists() {
        // Walk the mount point and copy relevant files to ~/Library/Audio/Plug-Ins/
        let walk_res = std::fs::read_dir(mount_point);
        if let Ok(entries) = walk_res {
            for entry in entries.flatten() {
                let path = entry.path();
                let ext = path.extension().and_then(|e| e.to_str()).unwrap_or("").to_lowercase();

                let target_dir = match ext.as_str() {
                    "vst" => "~/Library/Audio/Plug-Ins/VST",
                    "vst3" => "~/Library/Audio/Plug-Ins/VST3",
                    "component" => "~/Library/Audio/Plug-Ins/Components",
                    _ => continue,
                };

                let target_dir_expanded = target_dir.replace("~", &std::env::var("HOME").unwrap_or_default());

                // Ensure target directory exists
                let _ = std::fs::create_dir_all(&target_dir_expanded);

                let _ = Command::new("cp")
                    .args(&["-R", path.to_str().unwrap_or(""), &target_dir_expanded])
                    .output();

                copied_files.push(path.file_name().unwrap_or_default().to_string_lossy().into_owned());
            }
        }

        // Detach DMG
        let _ = Command::new("hdiutil")
            .args(&["detach", mount_point])
            .output();
    }

    if copied_files.is_empty() {
        Ok(format!("Mounted DMG but found no .vst/.vst3/.component files to copy: {}", file_path))
    } else {
        Ok(format!("Successfully copied from DMG: {:?}", copied_files))
    }
}

#[cfg(not(target_os = "macos"))]
fn install_dmg(_file_path: &str) -> Result<String, String> {
    Err("DMG installers are only supported on macOS".into())
}
