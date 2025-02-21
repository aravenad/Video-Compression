import subprocess
import threading
import re
import logging
import os
import shutil
import time
from tkinter import messagebox
from core.filename_utils import generate_unique_filename
from core.state import (
    active_compressions,
    compression_settings,
    est_size_label,
    create_ffmpeg_command,
    cleanup_current_file
)

# Configuration des logs
logging.basicConfig(
    filename="compression.log",
    filemode="w",
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)

# Verrou global pour éviter les conflits entre threads
lock = threading.Lock()

def update_progress(process, progress_bar, status_label, on_complete):
    """Met à jour la barre de progression et affiche les logs en temps réel."""
    try:
        duration = None
        last_update = time.time()
        start_time = time.time()

        while True:
            line = process.stdout.readline()
            if not line and process.poll() is not None:
                break
                
            line = line.strip()
            if not line:
                continue
            
            # Get duration once
            if not duration and "Duration" in line:
                match = re.search(r"Duration: (\d+):(\d+):(\d+)", line)
                if match:
                    h, m, s = map(int, match.groups())
                    duration = h * 3600 + m * 60 + s

            # Update progress
            if duration and "time=" in line:
                match = re.search(r"time=(\d+):(\d+):(\d+)", line)
                if match:
                    h, m, s = map(int, match.groups())
                    time_done = h * 3600 + m * 60 + s
                    progress = min(100, (time_done / duration) * 100)
                    
                    # Calculate remaining time
                    elapsed = time.time() - start_time
                    if progress > 0:
                        total = elapsed * (100 / progress)
                        remaining = total - elapsed
                        mins = int(remaining // 60)
                        secs = int(remaining % 60)
                        
                        # Update UI (immediately)
                        progress_bar["value"] = progress
                        progress_bar.update()
                        status_label.set(f"Temps restant: {mins}m {secs}s")

        if process.returncode == 0:
            progress_bar["value"] = 100
            progress_bar.update()
            status_label.set("Terminé")
            on_complete(False)
        else:
            cleanup_current_file()
            on_complete(True)

    except Exception as e:
        logging.error(f"Error in progress update: {e}")
        cleanup_current_file()
        on_complete(True)

MAX_FILE_SIZE = 20 * 1024 * 1024 * 1024  # 4 GB
COMPRESSION_TIMEOUT = 3600  # 1 heure

def compress_video(input_file, progress_bar, status_label, comp_frame=None):
    """
    Lance la compression dans un thread séparé.
    :param input_file: Fichier à compresser.
    :param progress_bar: Barre de progression.
    :param status_label: Label pour le statut.
    :param comp_frame: (Optionnel) UI frame associé à cette compression.
    """
    from core.state import compression_queue  # Ensure proper import
    with lock:
        # Removed limit on simultaneous compressions.
        if not input_file or input_file == "Aucun fichier sélectionné":
            raise ValueError("Aucun fichier sélectionné.")
        if os.path.getsize(input_file) > MAX_FILE_SIZE:
            raise ValueError("Fichier trop volumineux (max 4GB)")
        if shutil.disk_usage(os.path.dirname(input_file)).free < os.path.getsize(input_file):
            raise ValueError("Espace disque insuffisant")
        
        import core.state as state_module
        output_file = generate_unique_filename(input_file)
        state_module.current_output_file = output_file
        
        try:
            ffmpeg_command = create_ffmpeg_command(input_file, output_file)
        except RuntimeError as e:
            messagebox.showerror("Erreur", str(e))
            return

        def on_complete(canceled=False):
            if not canceled:
                state_module.compressed_files.append(input_file)
                final_size = os.path.getsize(output_file)
                original_size = os.path.getsize(input_file)
                reduction = ((original_size - final_size) / original_size) * 100
                messagebox.showinfo("Succès", 
                    f"Compression terminée : {output_file}\nRéduction: {reduction:.1f}%")
            state_module.current_output_file = None

        def run_compression():
            try:
                process = subprocess.Popen(
                    ffmpeg_command,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.STDOUT,
                    universal_newlines=True,
                    bufsize=1
                )
                active_compressions.append(process)
                if comp_frame is not None:
                    comp_frame.process = process
                try:
                    update_progress(process, progress_bar, status_label, 
                                    lambda canceled=False: on_complete(canceled))
                except subprocess.TimeoutExpired:
                    if state_module.current_output_file:
                        try:
                            os.remove(state_module.current_output_file)
                        except OSError:
                            pass
                    process.kill()
                    raise TimeoutError("La compression a pris trop de temps")
                finally:
                    try:
                        active_compressions.remove(process)
                    except ValueError:
                        pass

            except Exception as e:
                logging.error(f"Erreur : {e}")
                messagebox.showerror("Erreur", f"Une erreur est survenue : {e}")

        status_label.set("Compression en cours...")
        progress_bar["value"] = 0
        threading.Thread(target=run_compression, daemon=True).start()
