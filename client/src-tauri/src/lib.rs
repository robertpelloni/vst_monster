mod installer;

use tauri::{
    menu::{Menu, MenuItem},
    tray::{TrayIcon, TrayIconBuilder, TrayIconEvent},
    AppHandle, Manager,
};

#[tauri::command]
fn greet(name: &str) -> String {
    format!("Hello, {}! You've been greeted from Rust!", name)
}

#[tauri::command]
fn stop_servers() -> Result<String, String> {
    tauri::async_runtime::spawn(async {
        let client = reqwest::Client::new();
        let _ = client.post("http://localhost:3000/system/shutdown").send().await;
    });
    Ok("Servers stopping signal sent".to_string())
}

#[tauri::command]
fn quit_app(app: AppHandle) {
    tauri::async_runtime::spawn(async move {
        let client = reqwest::Client::new();
        let _ = client.post("http://localhost:3000/system/shutdown").send().await;
        app.exit(0);
    });
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .setup(|app| {
            // Build system tray icon menu
            let quit_i = MenuItem::with_id(app, "quit", "Quit VST Monster & Servers", true, None::<&str>)?;
            let show_i = MenuItem::with_id(app, "show", "Show Dashboard", true, None::<&str>)?;
            let hide_i = MenuItem::with_id(app, "hide", "Hide Dashboard", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show_i, &hide_i, &quit_i])?;

            let _tray = TrayIconBuilder::new()
                .icon(app.default_window_icon().unwrap().clone())
                .menu(&menu)
                .tooltip("VST Monster Desktop")
                .on_menu_event(|app: &AppHandle, event| match event.id.as_ref() {
                    "quit" => {
                        let app_clone = app.clone();
                        tauri::async_runtime::spawn(async move {
                            let client = reqwest::Client::new();
                            let _ = client.post("http://localhost:3000/system/shutdown").send().await;
                            app_clone.exit(0);
                        });
                    }
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    "hide" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.hide();
                        }
                    }
                    _ => {}
                })
                .on_tray_icon_event(|tray: &TrayIcon, event| {
                    if let TrayIconEvent::Click { .. } = event {
                        let app = tray.app_handle();
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                })
                .build(app)?;

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            greet,
            installer::install_plugin,
            installer::remove_plugin,
            installer::scan_installed_plugins,
            installer::get_default_scan_paths,
            installer::open_plugin_directory,
            installer::open_presets_directory,
            installer::verify_installed_plugin,
            stop_servers,
            quit_app
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
