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
    """
    Met à jour la barre de progression et affiche les logs en temps réel.
    """
    try:
        duration = None
        last_update_time = 0
        start_time = time.time()

        while True:
            output = process.stdout.readline()
            if not output and process.poll() is not None:
                break

            output = output.strip()
            logging.info(output)

            current_time = time.time()
            # Update UI maximum once every 100ms
            if current_time - last_update_time < 0.1:
                continue
            
            # Extraction de la durée totale
            if "Duration" in output and not duration:
                match = re.search(r"Duration: (\d+):(\d+):(\d+)\.(\d+)", output)
                if match:
                    hours, minutes, seconds, milliseconds = map(int, match.groups())
                    duration = (
                        hours * 3600 * 1000
                        + minutes * 60 * 1000
                        + seconds * 1000
                        + milliseconds * 10
                    )
                    logging.info(f"Durée totale (ms) : {duration}")

            # Extraction du temps traité
            if "out_time_us=" in output and duration:
                match = re.search(r"out_time_us=(\d+)", output)
                if match:
                    current_time = int(match.group(1)) // 1000  # ms
                    progress = min((current_time / duration) * 100, 100)
                    
                    # Calculate time estimation
                    elapsed_time = time.time() - start_time
                    if progress > 0:
                        total_estimated = elapsed_time * (100 / progress)
                        remaining_time = total_estimated - elapsed_time
                        
                        # Format remaining time
                        remaining_mins = int(remaining_time // 60)
                        remaining_secs = int(remaining_time % 60)
                        time_str = f"{remaining_mins}m {remaining_secs}s"
                        
                        status_label.set(f"Compression en cours... Temps restant: {time_str}")
                    
                    progress_bar["value"] = progress
                    progress_bar.update()

        # Fin de la progression
        if process.returncode == 0:  # Normal completion
            progress_bar["value"] = 100
            status_label.set("Compression terminée.")
            logging.info("Compression terminée.")
            on_complete(canceled=False)
        else:  # Process was terminated or failed
            logging.info("Compression annulée ou échouée.")
            cleanup_current_file()
            on_complete(canceled=True)
    except Exception as e:
        logging.error(f"Error in progress update: {e}")
        cleanup_current_file()
        on_complete(canceled=True)

    last_update_time = current_time

MAX_FILE_SIZE = 4 * 1024 * 1024 * 1024  # 4 GB
COMPRESSION_TIMEOUT = 3600  # 1 heure

def compress_video(input_file, progress_bar, status_label, comp_frame=None):
    """
    Lance la compression dans un thread séparé.
    :param input_file: Fichier à compresser.
    :param progress_bar: Barre de progression.
    :param status_label: Label pour le statut.
    :param comp_frame: (Optionnel) UI frame associé à cette compression.
    """
    with lock:
        if not input_file or input_file == "Aucun fichier sélectionné":
            raise ValueError("Aucun fichier sélectionné.")
        if os.path.getsize(input_file) > MAX_FILE_SIZE:
            raise ValueError("Fichier trop volumineux (max 4GB)")
        if shutil.disk_usage(os.path.dirname(input_file)).free < os.path.getsize(input_file):
            raise ValueError("Espace disque insuffisant")
        
        # Use a consistent module alias for state
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
