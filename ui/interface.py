import tkinter as tk
import os
import psutil
import logging
from tkinter import filedialog, messagebox, ttk
from core.compression import compress_video, MAX_FILE_SIZE
from core.state import (
    compression_queue, 
    MAX_QUEUE_SIZE, 
    active_compressions, 
    compression_settings
)

def start_app():
    root = tk.Tk()
    root.title("Compresseur Vidéo")
    
    # Initialize state
    process_status = tk.StringVar(value="En attente...")
    frame = tk.Frame(root, padx=20, pady=20)
    frame.pack()

    # Files list frame
    files_frame = ttk.LabelFrame(frame, text="Fichiers sélectionnés")
    files_frame.pack(fill="x", pady=10)
    files_listbox = tk.Listbox(files_frame, height=5, width=50)
    files_listbox.pack(pady=5)

    # Import state module
    import core.state
    
    # Progress frame with proper state management
    progress_frame = ttk.LabelFrame(frame, text="Progression")
    progress_frame.pack(fill="x", pady=10)
    
    stats_frame = ttk.Frame(progress_frame)
    stats_frame.pack(fill="x")
    
    est_size_label = ttk.Label(stats_frame, text="Taille estimée : --")
    est_size_label.pack(side=tk.RIGHT, padx=5)
    core.state.est_size_label = est_size_label
    
    memory_label = ttk.Label(stats_frame, text="Mémoire : 0%")
    memory_label.pack(side=tk.LEFT, padx=5)
    
    progress_bar = ttk.Progressbar(progress_frame, orient="horizontal", length=400, mode="determinate")
    progress_bar.pack(pady=5)
    core.state.progress_bar = progress_bar
    
    status_label = ttk.Label(progress_frame, textvariable=process_status)
    status_label.pack(pady=5)
    core.state.status_label = status_label

    # Settings frame
    settings_frame = ttk.LabelFrame(frame, text="Paramètres")
    settings_frame.pack(fill="x", pady=10)
    ttk.Label(settings_frame, text="Qualité :").grid(row=0, column=0, padx=5)
    quality_scale = ttk.Scale(settings_frame, from_=0, to=51, orient="horizontal")
    quality_scale.set(compression_settings['quality'])
    quality_scale.grid(row=0, column=1, padx=5)
    ttk.Label(settings_frame, text="Vitesse :").grid(row=1, column=0, padx=5)
    speed_combo = ttk.Combobox(settings_frame, 
                              values=['ultrafast', 'superfast', 'veryfast', 'faster', 
                                     'fast', 'medium', 'slow', 'slower', 'veryslow'])
    speed_combo.set(compression_settings['preset'])
    speed_combo.grid(row=1, column=1, padx=5)

    # Add delete-originals checkbox in Settings frame
    delete_original_var = tk.BooleanVar(value=False)
    delete_checkbox = ttk.Checkbutton(settings_frame, text="Supprimer fichiers originaux après compression", variable=delete_original_var)
    delete_checkbox.grid(row=2, column=0, columnspan=2, padx=5, pady=5)

    def update_settings(*args):
        compression_settings['quality'] = int(quality_scale.get())
        compression_settings['preset'] = speed_combo.get()

    quality_scale.configure(command=update_settings)
    speed_combo.bind('<<ComboboxSelected>>', update_settings)

    def select_files(single=False):
        if compression_queue.qsize() >= MAX_QUEUE_SIZE:
            messagebox.showwarning("Attention", "File d'attente pleine (max 10 fichiers)")
            return
            
        if single:
            files = filedialog.askopenfilename(
                title="Sélectionnez une vidéo",
                filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
            )
            files = [files] if files else []
        else:
            files = filedialog.askopenfilenames(
                title="Sélectionnez des vidéos",
                filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
            )
        
        valid_files = []
        for path in files:
            try:
                size = os.path.getsize(path)
                if size > MAX_FILE_SIZE:
                    messagebox.showwarning("Attention", f"Fichier trop volumineux: {path}")
                    continue
                valid_files.append(path)
                files_listbox.insert(tk.END, os.path.basename(path))
                compression_queue.put(path)
            except OSError as e:
                messagebox.showerror("Erreur", f"Erreur d'accès au fichier: {path}")
                
        if valid_files:
            process_status.set(f"{len(valid_files)} fichiers ajoutés à la file.")

    def process_queue():
        if compression_queue.empty():
            process_status.set("File d'attente vide")
            return

        if active_compressions:
            process_status.set("Compression en cours...")
            return

        next_file = compression_queue.get()
        try:
            # Add debug logging
            logging.info(f"Starting compression for: {next_file}")
            compress_video(next_file, progress_bar, process_status)  # Changed from status_label
            files_listbox.delete(0)
            
            # Schedule next file processing
            if not compression_queue.empty():
                root.after(1000, process_queue)
        except Exception as e:
            logging.error(f"Compression error: {e}")
            messagebox.showerror("Erreur", str(e))
            process_status.set("Erreur de compression")

    def update_queue_status():
        queue_size = compression_queue.qsize()
        if queue_size > 0:
            process_status.set(f"{queue_size} fichiers en attente")
        root.after(500, update_queue_status)

    # Start queue status updates
    update_queue_status()
    
    def cancel_compression():
        if not active_compressions:
            return

        # Store current file path before cleanup
        import core.state
        current_file = core.state.current_output_file

        for process in active_compressions[:]:
            if process and process.poll() is None:
                try:
                    process.terminate()
                    process.wait(timeout=1)
                    if current_file and os.path.exists(current_file):
                        os.remove(current_file)
                        logging.info(f"Deleted incomplete file: {current_file}")
                except Exception as e:
                    logging.error(f"Error during compression cancellation: {e}")
        
        active_compressions.clear()
        core.state.current_output_file = None
        process_status.set("Compression annulée")
        progress_bar["value"] = 0

    # Buttons frame (moved up)
    buttons_frame = ttk.Frame(frame)
    buttons_frame.pack(fill="x", pady=10)
    
    select_button = ttk.Button(buttons_frame, text="Sélectionner des fichiers", 
                              command=lambda: select_files(False))
    select_button.pack(side=tk.LEFT, padx=5)
    
    compress_button = ttk.Button(buttons_frame, text="Compresser", command=process_queue)
    compress_button.pack(side=tk.LEFT, padx=5)
    
    cancel_button = ttk.Button(buttons_frame, text="Annuler", command=cancel_compression)
    cancel_button.pack(side=tk.RIGHT, padx=5)
    cancel_button["state"] = "disabled"  # Initially disabled

    def update_buttons():
        if active_compressions:
            cancel_button["state"] = "normal"
            compress_button["state"] = "disabled"
        else:
            cancel_button["state"] = "disabled"
            compress_button["state"] = "normal"
        root.after(100, update_buttons)

    # Start button state updates after buttons are created
    root.after(100, update_buttons)
    
    update_buttons()

    # Update state references
    import core.state
    core.state.est_size_label = est_size_label
    core.state.progress_bar = progress_bar
    core.state.status_label = status_label

    def update_resource_monitor():
        try:
            memory_percent = psutil.Process().memory_percent()
            memory_label.config(text=f"Mémoire : {memory_percent:.1f}%")
            root.after(1000, update_resource_monitor)
        except Exception as e:
            logging.error(f"Resource monitoring error: {e}")

    root.after(1000, update_resource_monitor)

    # New: Periodically check and delete original files once compression is done and checkbox is checked.
    def check_delete_originals():
        from core import state as state_module
        if compression_queue.empty() and not active_compressions and delete_original_var.get():
            # Delete only files that were in the compression queue and recorded as compressed
            for original in state_module.compressed_files:
                try:
                    os.remove(original)
                    logging.info(f"Deleted original file: {original}")
                except Exception as e:
                    logging.error(f"Error deleting {original}: {e}")
            state_module.compressed_files.clear()
            process_status.set("Fichiers originaux supprimés.")
        root.after(1000, check_delete_originals)

    # Start queue status updates and deletion check
    update_queue_status()
    root.after(1000, check_delete_originals)

    root.mainloop()
