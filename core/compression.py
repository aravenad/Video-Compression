import subprocess
import threading
import re
import logging
from tkinter import messagebox
from core.filename_utils import generate_unique_filename

# Configuration des logs
logging.basicConfig(
    filename="compression.log",
    filemode="w",
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)

def update_progress(process, progress_bar, status_label, on_complete):
    """
    Met à jour la barre de progression et affiche les logs en temps réel.
    """
    duration = None

    while True:
        output = process.stdout.readline()
        if not output and process.poll() is not None:
            break

        output = output.strip()
        logging.info(output)

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
                current_time = int(match.group(1)) // 1000  # Conversion en ms
                progress = min((current_time / duration) * 100, 100)
                logging.info(f"Temps actuel (ms) : {current_time}, Progression : {progress:.2f}%")

                # Mise à jour de la barre de progression
                progress_bar["value"] = progress
                progress_bar.update()

    # Fin de la progression
    progress_bar["value"] = 100
    status_label.set("Compression terminée.")
    logging.info("Compression terminée.")
    on_complete()

def compress_video(input_file, progress_bar, status_label):
    """
    Lance la compression dans un thread séparé.
    :param input_file: Fichier à compresser.
    :param progress_bar: Barre de progression.
    :param status_label: Label pour le statut.
    """
    if not input_file or input_file == "Aucun fichier sélectionné":
        raise ValueError("Aucun fichier sélectionné.")

    output_file = generate_unique_filename(input_file)
    ffmpeg_command = [
        "ffmpeg", "-i", input_file,
        "-vcodec", "libx264", "-crf", "23",
        "-y", "-progress", "pipe:1", output_file
    ]

    def on_complete():
        messagebox.showinfo("Succès", f"Compression terminée : {output_file}")

    def run_compression():
        try:
            process = subprocess.Popen(
                ffmpeg_command,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1
            )
            update_progress(process, progress_bar, status_label, on_complete)
        except Exception as e:
            logging.error(f"Erreur : {e}")
            messagebox.showerror("Erreur", f"Une erreur est survenue : {e}")

    # Démarrage du processus
    status_label.set("Compression en cours...")
    progress_bar["value"] = 0
    threading.Thread(target=run_compression, daemon=True).start()
