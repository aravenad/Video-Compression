import tkinter as tk
from tkinter import ttk, filedialog, messagebox
import os
import psutil
import logging
import GPUtil
from core.compression import compress_video, MAX_FILE_SIZE
from core.state import compression_queue, MAX_QUEUE_SIZE, active_compressions, compression_settings, pending_files

def build_file_queue_tab(notebook, queue_status_var):
    frame = ttk.Frame(notebook)
    notebook.add(frame, text="File d'attente")
    
    # Pending files list
    files_frame = ttk.LabelFrame(frame, text="Fichiers sélectionnés")
    files_frame.pack(fill="x", pady=10, padx=5)
    files_listbox = tk.Listbox(files_frame, height=8)
    files_listbox.pack(side="left", fill="x", expand=True, padx=5, pady=5)
    files_scroll = ttk.Scrollbar(files_frame, orient="vertical", command=files_listbox.yview)
    files_listbox.configure(yscrollcommand=files_scroll.set)
    files_scroll.pack(side="right", fill="y")
    
    # Delete selected file button with right margin
    def delete_selected_file():
        sel = files_listbox.curselection()
        if not sel:
            return
        index = sel[0]
        if index < len(pending_files):
            pending_files.pop(index)
        files_listbox.delete(index)
        # Rebuild the compression_queue from pending_files.
        while not compression_queue.empty():
            compression_queue.get_nowait()
        for f in pending_files:
            compression_queue.put(f)
        queue_status_var.set(f"{compression_queue.qsize()} fichiers en attente")
    del_button = ttk.Button(files_frame, text="Supprimer", command=delete_selected_file)
    del_button.pack(pady=5, padx=(5, 15))  # Added right padding
    
    # Clear Queue button with right margin
    def clear_queue():
        pending_files.clear()
        files_listbox.delete(0, tk.END)
        while not compression_queue.empty():
            compression_queue.get_nowait()
        queue_status_var.set("0 fichiers en attente")
    clear_button = ttk.Button(files_frame, text="Nettoyer", command=clear_queue)
    clear_button.pack(pady=5, padx=(5, 15))  # Added right padding
    
    # File selection button
    def select_files(single=False):
        if compression_queue.qsize() >= MAX_QUEUE_SIZE:
            messagebox.showwarning("Attention", "File d'attente pleine")
            return
        files = filedialog.askopenfilenames(
                    title="Sélectionnez des vidéos",
                    filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")])
        for path in files:
            try:
                if os.path.getsize(path) > MAX_FILE_SIZE:
                    messagebox.showwarning("Attention", f"Fichier trop volumineux : {path}")
                    continue
                files_listbox.insert(tk.END, os.path.basename(path))
                compression_queue.put(path)
                pending_files.append(path)
            except OSError:
                messagebox.showerror("Erreur", f"Erreur d'accès à {path}")
        queue_status_var.set(f"{compression_queue.qsize()} fichiers en attente")
    sel_button = ttk.Button(frame, text="Sélectionner des fichiers", command=lambda: select_files(False))
    sel_button.pack(pady=5)
    
    return files_listbox

def build_compression_tab(notebook):
    tab = ttk.Frame(notebook)
    notebook.add(tab, text="Compressions en cours")
    
    # Scrollable frame for active compressions
    run_frame = ttk.LabelFrame(tab, text="Compresions en cours")
    run_frame.pack(fill="both", expand=True, padx=5, pady=5)
    
    # Configure canvas with scrollbar
    canvas = tk.Canvas(run_frame, height=200)  # Remove fixed width
    canvas.pack(side="left", fill="both", expand=True, padx=(5, 0))  # Add left padding
    
    vbar = ttk.Scrollbar(run_frame, orient="vertical", command=canvas.yview)
    vbar.pack(side="right", fill="y")
    
    canvas.configure(yscrollcommand=vbar.set)
    
    # Inner frame to hold compression frames
    inner_frame = ttk.Frame(canvas)
    canvas.create_window((0, 0), window=inner_frame, anchor="nw", width=canvas.winfo_width())
    
    # Update inner frame width when canvas size changes
    def on_canvas_configure(event):
        canvas.itemconfig(canvas.find_withtag("all")[0], width=event.width)
        
    canvas.bind("<Configure>", on_canvas_configure)
    inner_frame.bind("<Configure>", lambda e: canvas.configure(scrollregion=canvas.bbox("all")))
    
    return inner_frame

def build_settings_tab(notebook):
    settings = ttk.Frame(notebook)
    notebook.add(settings, text="Paramètres")
    
    # Quality and Preset controls
    ttk.Label(settings, text="Qualité :").grid(row=0, column=0, padx=5, pady=5, sticky="w")
    quality_scale = ttk.Scale(settings, from_=0, to=51, orient="horizontal")
    quality_scale.set(compression_settings['quality'])
    quality_scale.grid(row=0, column=1, padx=5, pady=5, sticky="ew")
    
    ttk.Label(settings, text="Vitesse :").grid(row=1, column=0, padx=5, pady=5, sticky="w")
    speed_combo = ttk.Combobox(settings, values=['ultrafast','superfast','veryfast','faster','fast','medium','slow','slower','veryslow'])
    speed_combo.set(compression_settings['preset'])
    speed_combo.grid(row=1, column=1, padx=5, pady=5, sticky="ew")
    
    def update_settings(*args):
        compression_settings['quality'] = int(quality_scale.get())
        compression_settings['preset'] = speed_combo.get()
    quality_scale.configure(command=update_settings)
    speed_combo.bind('<<ComboboxSelected>>', update_settings)
    
    # Delete originals checkbox
    delete_original_var = tk.BooleanVar(value=False)
    ttk.Checkbutton(settings, text="Supprimer les fichiers originaux après compression", variable=delete_original_var)\
        .grid(row=2, column=0, columnspan=2, padx=5, pady=5, sticky="w")
    
    return None  # Settings state can be stored globally if needed.

def build_performance_tab(notebook):
    perf_frame = ttk.Frame(notebook)
    notebook.add(perf_frame, text="Performance")
    
    # Memory Usage
    ttk.Label(perf_frame, text="Mémoire :").grid(row=0, column=0, padx=5, pady=5, sticky="e")
    memory_progress = ttk.Progressbar(perf_frame, length=200, mode="determinate")
    memory_progress.grid(row=0, column=1, padx=5, pady=5)
    memory_label = ttk.Label(perf_frame, text="0%")
    memory_label.grid(row=0, column=2, padx=5, pady=5)
    
    # CPU Usage
    ttk.Label(perf_frame, text="CPU :").grid(row=1, column=0, padx=5, pady=5, sticky="e")
    cpu_progress = ttk.Progressbar(perf_frame, length=200, mode="determinate")
    cpu_progress.grid(row=1, column=1, padx=5, pady=5)
    cpu_label = ttk.Label(perf_frame, text="0%")
    cpu_label.grid(row=1, column=2, padx=5, pady=5)
    
    # GPU Usage
    ttk.Label(perf_frame, text="GPU :").grid(row=2, column=0, padx=5, pady=5, sticky="e")
    gpu_progress = ttk.Progressbar(perf_frame, length=200, mode="determinate")
    gpu_progress.grid(row=2, column=1, padx=5, pady=5)
    gpu_label = ttk.Label(perf_frame, text="0%")
    gpu_label.grid(row=2, column=2, padx=5, pady=5)
    
    def update_perf():
        try:
            # Update Memory
            memory_percent = psutil.Process().memory_percent()
            memory_progress["value"] = memory_percent
            memory_label.config(text=f"{memory_percent:.1f}%")
            
            # Update CPU
            cpu_usage = psutil.cpu_percent(interval=None)
            cpu_progress["value"] = cpu_usage
            cpu_label.config(text=f"{cpu_usage:.1f}%")
            
            # Update GPU if available
            gpus = GPUtil.getGPUs()
            if gpus:
                gpu_usage = max(gpu.load * 100 for gpu in gpus)
                gpu_progress["value"] = gpu_usage
                gpu_label.config(text=f"{gpu_usage:.1f}%")
        except Exception as e:
            logging.error(f"Performance monitoring error: {e}")
        
        perf_frame.after(1000, update_perf)
    
    perf_frame.after(1000, update_perf)
    return perf_frame

def start_app():
    root = tk.Tk()
    root.title("Compresseur Vidéo - v2.0")
    notebook = ttk.Notebook(root)
    notebook.pack(fill="both", expand=True, padx=10, pady=10)
    
    # Add style for selected compression frame
    style = ttk.Style()
    style.configure("Selected.TFrame", background="lightblue")
    
    # Add container for selected compression frame
    selected_running_comp = [None]
    
    def select_running_comp(frame):
        if selected_running_comp[0]:
            selected_running_comp[0].configure(style="TFrame")
        selected_running_comp[0] = frame
        frame.configure(style="Selected.TFrame")
    
    queue_status_var = tk.StringVar(value="En attente...")
    files_listbox = build_file_queue_tab(notebook, queue_status_var)
    active_frame = build_compression_tab(notebook)
    build_settings_tab(notebook)
    build_performance_tab(notebook)
    
    # Status bar at bottom
    status_bar = ttk.Label(root, textvariable=queue_status_var, relief="sunken", anchor="w")
    status_bar.pack(side="bottom", fill="x")
    
    # Global buttons frame for compression and cancellation.
    btn_frame = ttk.Frame(root)
    btn_frame.pack(fill="x", padx=10, pady=5)
    compress_button = ttk.Button(
        btn_frame, 
        text="Compresser", 
        command=lambda: process_queue(active_frame, root, queue_status_var, select_running_comp)
    )
    compress_button.pack(side="left", padx=5)
    cancel_button = ttk.Button(
        btn_frame, 
        text="Annuler", 
        command=lambda: cancel_active(selected_running_comp, queue_status_var)
    )
    cancel_button.pack(side="right", padx=5)
    
    # Monitor active compressions: remove finished frames and keep compress_button enabled.
    def monitor_active():
        active_count = len(active_compressions)
        for child in active_frame.winfo_children():
            proc = getattr(child, "process", None)
            if proc and proc.poll() is not None:
                child.destroy()
        
        # Update status based on number of active compressions
        if active_count == 0:
            queue_status_var.set("En attente...")
        elif active_count == 1:
            # Find the name of the only active compression
            for child in active_frame.winfo_children():
                if hasattr(child, 'process'):
                    name = child.winfo_children()[0].cget("text").replace("En cours : ", "")
                    queue_status_var.set(f"Compression de {name}...")
                    break
        else:
            queue_status_var.set(f"Multiple compressions en cours ({active_count})...")
        
        root.after(1000, monitor_active)
    
    monitor_active()
    
    root.mainloop()

def process_queue(active_frame, root, queue_status_var, select_running_comp):
    """Process all files in queue automatically"""
    from core.state import compression_queue
    if compression_queue.empty():
        return

    next_file = compression_queue.get()
    # Update status message for the new compression
    active_count = len(active_compressions)
    if active_count == 0:
        queue_status_var.set(f"Compression de {os.path.basename(next_file)}...")
    else:
        queue_status_var.set(f"Multiple compressions en cours ({active_count + 1})...")

    # Create compression frame and start compression
    comp_frame = ttk.Frame(active_frame, relief="sunken", borderwidth=1, padding=5)
    comp_frame.pack(fill="x", pady=3, padx=0)  # Remove horizontal padding
    # Add click binding
    comp_frame.bind("<Button-1>", lambda e: select_running_comp(comp_frame))
    
    file_label = ttk.Label(comp_frame, text=f"En cours: {os.path.basename(next_file)}")
    file_label.pack(anchor="w", padx=2, pady=2)
    
    progress_bar = ttk.Progressbar(comp_frame, orient="horizontal", length=400, mode="determinate")
    progress_bar.pack(pady=2)
    status_var = tk.StringVar(value="Démarrage...")
    status_label = ttk.Label(comp_frame, textvariable=status_var)
    status_label.pack(anchor="w", pady=2)
    
    # Launch compression
    compress_video(next_file, progress_bar, status_var, comp_frame)
    
    if not compression_queue.empty():
        root.after(1000, lambda: process_queue(active_frame, root, queue_status_var, select_running_comp))

def cancel_active(selected_running_comp, queue_status_var):
    """Cancel the selected compression and cleanup its resources."""
    # Get the currently selected frame
    if not selected_running_comp[0]:
        messagebox.showwarning("Attention", "Veuillez sélectionner une compression à annuler.")
        return

    comp_frame = selected_running_comp[0]
    process = getattr(comp_frame, "process", None)
    
    if not process or process.poll() is not None:
        messagebox.showinfo("Info", "Cette compression est déjà terminée.")
        return

    try:
        # Get the filename being compressed for the status message
        filename = comp_frame.winfo_children()[0].cget("text").replace("En cours: ", "")
        
        # Kill the ffmpeg process
        process.terminate()
        process.wait(timeout=1)
        
        # Remove from active compressions list
        if process in active_compressions:
            active_compressions.remove(process)
        
        # Delete the incomplete output file
        import core.state as state_module
        if state_module.current_output_file and os.path.exists(state_module.current_output_file):
            os.remove(state_module.current_output_file)
            logging.info(f"Deleted incomplete file: {state_module.current_output_file}")
        
        # Remove the compression frame from UI
        comp_frame.destroy()
        selected_running_comp[0] = None
        
        # Update status
        queue_status_var.set(f"Compression de {filename} annulée.")
        
    except Exception as e:
        logging.error(f"Error during compression cancellation: {e}")
        messagebox.showerror("Erreur", f"Erreur lors de l'annulation : {e}")

if __name__ == "__main__":
    start_app()
