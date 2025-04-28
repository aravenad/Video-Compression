// Prevents additional console window on Windows in release, DO NOT REMOVE!!
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use serde::{Deserialize, Serialize};
use std::fs;
use std::path::Path;
use tauri::Manager;

#[derive(Debug, Serialize, Deserialize)]
struct PresetInfo {
    video_codec: String,
    crf: u32,
    preset: String,
    description: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct ConfigFile {
    presets: std::collections::HashMap<String, PresetInfo>,
}

#[tauri::command]
fn get_presets() -> Result<Vec<serde_json::Value>, String> {
    // Path to the config file
    let config_path = Path::new("C:\\Users\\xodai\\Documents\\Dev\\Video-Compression\\config\\default.yaml");
    
    // Read the config file
    let config_content = fs::read_to_string(config_path)
        .map_err(|e| format!("Failed to read config file: {}", e))?;
    
    // Parse the YAML
    let config: ConfigFile = serde_yaml::from_str(&config_content)
        .map_err(|e| format!("Failed to parse YAML: {}", e))?;
    
    // Convert presets to a list of objects with name and details
    let mut presets = Vec::new();
    for (name, info) in config.presets {
        let preset = serde_json::json!({
            "value": name,
            "label": name.split('-').map(|s| {
                let mut c = s.chars();
                c.next().map(|f| f.to_uppercase().collect::<String>()).unwrap_or_default() + &s[1..]
            }).collect::<Vec<_>>().join(" "), // Convert "preset-name" to "Preset Name"
            "description": info.description,
        });
        presets.push(preset);
    }
    
    Ok(presets)
}

#[tauri::command]
fn run_cli(args: Vec<String>) -> Result<String, String> {
  let mut cmd = std::process::Command::new("video-compress");
  cmd.args(&args);
  let output = cmd.output().map_err(|e| e.to_string())?;
  if !output.status.success() {
    return Err(String::from_utf8_lossy(&output.stderr).into_owned());
  }
  Ok(String::from_utf8_lossy(&output.stdout).into_owned())
}

fn main() {
    // gui_lib::run()  // Remove or comment out if not needed, as it may conflict with Tauri's main loop.
    tauri::Builder::default()
      .invoke_handler(tauri::generate_handler![run_cli, get_presets])
      .setup(|app| {
        let window = app.get_webview_window("main").unwrap();
        window.set_size(tauri::Size::Logical(tauri::LogicalSize {
          width: 1100.0,
          height: 800.0,
        })).ok();
        Ok(())
      })
      .run(tauri::generate_context!())
      .expect("error while running tauri application");
}
