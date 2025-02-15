import tkinter as tk
import os
import psutil
import logging
import GPUtil  # For GPU monitoring
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
    
    # Create separate variables: one for time left, one for queue/status messages.
    time_left_var = tk.StringVar(value="Temps restant: --")
    queue_status_var = tk.StringVar(value="En attente...")

    frame = tk.Frame(root, padx=20, pady=20)
    frame.pack()
    
    # Files list frame with scrollbar
    files_frame = ttk.LabelFrame(frame, text="Fichiers sélectionnés")
    files_frame.pack(fill="x", pady=10)
    files_listbox = tk.Listbox(files_frame, height=5, width=50)
    files_listbox.pack(side="left", pady=5, fill="x", expand=True)
    files_scroll = ttk.Scrollbar(files_frame, orient="vertical", command=files_listbox.yview)
    files_listbox.configure(yscrollcommand=files_scroll.set)
    files_scroll.pack(side="right", fill="y")
    # New: Button to delete selected file from list and queue
    def delete_selected_file():
        sel = files_listbox.curselection()
        if not sel:
            return
        index = sel[0]
        import core.state as state_module
        # Remove from pending_files list based on order in UI.
        if index < len(state_module.pending_files):
            state_module.pending_files.pop(index)
        files_listbox.delete(index)
        # Rebuild the compression_queue from pending_files.
        while not compression_queue.empty():
            compression_queue.get_nowait()
        for f in state_module.pending_files:
            compression_queue.put(f)
        queue_status_var.set(f"{compression_queue.qsize()} fichiers en attente")
    delete_button = ttk.Button(files_frame, text="Supprimer du queue", command=delete_selected_file)
    delete_button.pack(side="bottom", pady=5)
    
    # Rename 'Progression' frame to 'Performances' and show only performance stats
    performances_frame = ttk.LabelFrame(frame, text="Performances")
    performances_frame.pack(fill="x", pady=10)
    
    stats_frame = ttk.Frame(performances_frame)
    stats_frame.pack(fill="x")
    
    memory_label = ttk.Label(stats_frame, text="Mémoire : 0%")
    memory_label.pack(side=tk.LEFT, padx=5)
    cpu_label = ttk.Label(stats_frame, text="CPU : 0%")
    cpu_label.pack(side=tk.LEFT, padx=5)
    gpu_label = ttk.Label(stats_frame, text="GPU : 0%")
    gpu_label.pack(side=tk.LEFT, padx=5)
    
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
    
    # Delete-originals checkbox
    delete_original_var = tk.BooleanVar(value=False)
    delete_checkbox = ttk.Checkbutton(settings_frame, 
                                      text="Supprimer fichiers originaux après compression", 
                                      variable=delete_original_var)
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
        import core.state as state_module
        for path in files:
            try:
                size = os.path.getsize(path)
                if size > MAX_FILE_SIZE:
                    messagebox.showwarning("Attention", f"Fichier trop volumineux: {path}")
                    continue
                valid_files.append(path)
                files_listbox.insert(tk.END, os.path.basename(path))
                compression_queue.put(path)
                state_module.pending_files.append(path)
            except OSError as e:
                messagebox.showerror("Erreur", f"Erreur d'accès au fichier: {path}")
        if valid_files:
            queue_status_var.set(f"{len(valid_files)} fichiers ajoutés à la file.")
    
    # Replace existing running_frame with a scrollable frame for running compressions
    running_labelframe = ttk.LabelFrame(frame, text="Compresions en cours")
    running_labelframe.pack(fill="both", pady=10)
    # Increased width to 500 to show the full progress bar even with a scrollbar.
    scroll_canvas = tk.Canvas(running_labelframe, height=200, width=500)
    scrollbar = ttk.Scrollbar(running_labelframe, orient="vertical", command=scroll_canvas.yview)
    scroll_canvas.configure(yscrollcommand=scrollbar.set)
    scrollbar.pack(side="right", fill="y")
    scroll_canvas.pack(side="left", fill="both", expand=True)
    running_frame = ttk.Frame(scroll_canvas)
    scroll_canvas.create_window((0,0), window=running_frame, anchor="nw")
    
    def on_running_frame_config(event):
        scroll_canvas.configure(scrollregion=scroll_canvas.bbox("all"))
    running_frame.bind("<Configure>", on_running_frame_config)

    # Mutable container to store the selected running compression frame
    selected_running_comp = [None]
    
    def select_running_comp(frame):
        # Deselect previous frame if exists
        if selected_running_comp[0] is not None:
            selected_running_comp[0].configure(style="TFrame")
        selected_running_comp[0] = frame
        frame.configure(style="Selected.TFrame")
    
    # (Optional) Define a style for selected running frame
    style = ttk.Style()
    style.configure("Selected.TFrame", background="lightblue")
    
    def process_queue():
        if compression_queue.empty():
            queue_status_var.set("File d'attente vide")
            return
        next_file = compression_queue.get()
        try:
            logging.info(f"Starting compression for: {next_file}")
            # Create a new subframe for this compression
            comp_frame = ttk.Frame(running_frame, relief="sunken", borderwidth=1, padding=5)
            comp_frame.pack(fill="x", pady=3)
            # NEW: Adjust the frame width to match the scroll_canvas width
            comp_frame.config(width=scroll_canvas.winfo_width())
            comp_frame.bind("<Button-1>", lambda event, frame=comp_frame: select_running_comp(frame))
            file_label = ttk.Label(comp_frame, text=f"Compression: {os.path.basename(next_file)}")
            file_label.pack(anchor="w")
            local_time_var = tk.StringVar(value="Temps restant: --")
            local_status = ttk.Label(comp_frame, textvariable=local_time_var)
            local_status.pack(anchor="w", pady=2)
            local_progress = ttk.Progressbar(comp_frame, orient="horizontal", length=400, mode="determinate")
            local_progress.pack(pady=2)
            # Launch compression with its dedicated widgets and pass the comp_frame
            compress_video(next_file, local_progress, local_time_var, comp_frame)
            files_listbox.delete(0)
            if not compression_queue.empty():
                root.after(1000, process_queue)
        except Exception as e:
            logging.error(f"Compression error: {e}")
            messagebox.showerror("Erreur", str(e))
            queue_status_var.set("Erreur de compression")
    
    def cancel_compression():
        # Cancel only the selected running compression
        if selected_running_comp[0] is None:
            return
        comp = selected_running_comp[0]
        process = getattr(comp, "process", None)
        # Fix: use proper import alias for core.state
        import core.state as state_module
        if process and process.poll() is None:
            try:
                process.terminate()
                process.wait(timeout=1)
                # Delete the output file if exists
                if state_module.current_output_file and os.path.exists(state_module.current_output_file):
                    os.remove(state_module.current_output_file)
                    logging.info(f"Deleted incomplete file: {state_module.current_output_file}")
            except Exception as e:
                logging.error(f"Error during compression cancellation: {e}")
        comp.destroy()
        selected_running_comp[0] = None
        queue_status_var.set("Compression annulée")
    
    buttons_frame = ttk.Frame(frame)
    buttons_frame.pack(fill="x", pady=10)
    
    select_button = ttk.Button(buttons_frame, text="Sélectionner des fichiers", 
                               command=lambda: select_files(False))
    select_button.pack(side=tk.LEFT, padx=5)
    
    compress_button = ttk.Button(buttons_frame, text="Compresser", command=process_queue)
    compress_button.pack(side=tk.LEFT, padx=5)
    
    cancel_button = ttk.Button(buttons_frame, text="Annuler", command=cancel_compression)
    cancel_button.pack(side=tk.RIGHT, padx=5)
    cancel_button["state"] = "disabled"
    
    def update_buttons():
        if active_compressions:
            cancel_button["state"] = "normal"
            compress_button["state"] = "disabled"
        else:
            cancel_button["state"] = "disabled"
            compress_button["state"] = "normal"
        root.after(100, update_buttons)
    
    root.after(100, update_buttons)
    
    def update_queue_status():
        qs = f"{compression_queue.qsize()} fichiers en attente" if compression_queue.qsize() else "En attente..."
        queue_status_var.set(qs)
        root.after(500, update_queue_status)
    
    update_queue_status()
    
    def update_resource_monitor():
        try:
            memory_percent = psutil.Process().memory_percent()
            memory_label.config(text=f"Mémoire : {memory_percent:.1f}%")
            cpu_usage = psutil.cpu_percent(interval=None)
            cpu_label.config(text=f"CPU : {cpu_usage:.1f}%")
            gpus = GPUtil.getGPUs()
            if gpus:
                gpu_usage = max(gpu.load for gpu in gpus) * 100
                gpu_label.config(text=f"GPU : {gpu_usage:.1f}%")
            else:
                gpu_label.config(text="GPU : 0%")
            root.after(1000, update_resource_monitor)
        except Exception as e:
            logging.error(f"Resource monitoring error: {e}")
    
    root.after(1000, update_resource_monitor)
    
    def check_delete_originals():
        from core import state as state_module
        if compression_queue.empty() and not active_compressions and delete_original_var.get():
            for original in state_module.compressed_files:
                try:
                    os.remove(original)
                    logging.info(f"Deleted original file: {original}")
                except Exception as e:
                    logging.error(f"Error deleting {original}: {e}")
            state_module.compressed_files.clear()
            queue_status_var.set("Fichiers originaux supprimés.")
        root.after(1000, check_delete_originals)
    
    root.after(1000, check_delete_originals)
    
    root.mainloop()

if __name__ == "__main__":
    start_app()
