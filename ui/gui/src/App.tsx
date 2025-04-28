import { invoke } from "@tauri-apps/api/core";
import { open } from "@tauri-apps/plugin-dialog";
import { useEffect, useState } from "react";
import "./App.css";

type PresetOption = {
  value: string;
  label: string;
  description: string;
};

function App() {
  const [output, setOutput] = useState<string>("");
  const [files, setFiles] = useState<string[]>([]);
  const [preset, setPreset] = useState<string>("default");
  const [jobs, setJobs] = useState<number>(1);
  const [loading, setLoading] = useState<boolean>(false);
  const [presetOptions, setPresetOptions] = useState<PresetOption[]>([]);

  // Load presets from YAML config file when component mounts
  useEffect(() => {
    async function loadPresets() {
      try {
        const presets = await invoke<PresetOption[]>("get_presets");
        setPresetOptions(presets);
      } catch (error) {
        console.error("Failed to load presets:", error);
        setOutput(`Error loading presets: ${error}`);
      }
    }

    loadPresets();
  }, []);

  async function pickFiles() {
    const selected = (await open({
      multiple: true,
      directory: false,
      filters: [{ name: "Videos", extensions: ["mp4", "mov", "avi", "mkv"] }],
    })) as string[];
    if (selected) setFiles(Array.isArray(selected) ? selected : [selected]);
  }

  async function runCompress() {
    setOutput("");
    setLoading(true);
    const args = [
      "compress",
      "--jobs",
      jobs.toString(),
      "--preset",
      preset,
      ...files,
    ];
    try {
      const result: string = await invoke("run_cli", { args });
      setOutput(result);
    } catch (err) {
      setOutput("Error: " + String(err));
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Video Compressor</h1>
      </header>

      <main className="app-content">
        <section className="control-panel">
          <div className="card">
            <h2 className="card-title">Compression Settings</h2>

            <div className="control-group">
              <label className="control-label">Input Files</label>
              <button className="primary-button" onClick={pickFiles}>
                Select Videos
              </button>
              <div className="file-list">
                {files.length > 0 ? (
                  <ul>
                    {files.map((file, index) => (
                      <li key={index} className="file-item">
                        {file.split("\\").pop()}
                      </li>
                    ))}
                  </ul>
                ) : (
                  <div className="empty-state">No files selected</div>
                )}
              </div>
            </div>

            <div className="controls-row">
              <div className="control-group">
                <label className="control-label" htmlFor="preset-select">
                  Preset
                </label>
                <select
                  id="preset-select"
                  className="select-control"
                  value={preset}
                  onChange={(e) => setPreset(e.currentTarget.value)}
                >
                  {presetOptions.map((option) => (
                    <option
                      key={option.value}
                      value={option.value}
                      title={option.description}
                    >
                      {option.label}
                    </option>
                  ))}
                </select>
                {presetOptions.find((p) => p.value === preset)?.description && (
                  <div className="preset-description">
                    {presetOptions.find((p) => p.value === preset)?.description}
                  </div>
                )}
              </div>

              <div className="control-group">
                <label className="control-label" htmlFor="jobs-input">
                  Parallel Jobs
                </label>
                <input
                  id="jobs-input"
                  className="input-control"
                  type="number"
                  min="1"
                  value={jobs}
                  onChange={(e) => setJobs(Number(e.currentTarget.value))}
                />
              </div>
            </div>

            <div className="action-row">
              <button
                className={`primary-button compress-button ${
                  loading ? "loading" : ""
                }`}
                onClick={runCompress}
                disabled={files.length === 0 || loading}
              >
                {loading ? "Compressing..." : "Start Compression"}
              </button>
            </div>
          </div>
        </section>

        <section className="output-panel">
          <div className="card">
            <h2 className="card-title">Console Output</h2>
            <div className="console">
              {output ? (
                <pre className="console-text">{output}</pre>
              ) : (
                <div className="empty-state">Output will appear here</div>
              )}
            </div>
          </div>
        </section>
      </main>

      <footer className="app-footer">
        <p>Powered by Tauri and React</p>
      </footer>
    </div>
  );
}

export default App;
