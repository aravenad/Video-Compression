import tkinter as tk
from tkinter import filedialog, messagebox, ttk
import subprocess
import re
import threading

def select_file():
    """
    Ouvre une boîte de dialogue pour sélectionner un fichier vidéo.
    Met à jour le label pour afficher le chemin du fichier sélectionné.
    :return: None
    """
    file_path = filedialog.askopenfilename(
        title="Sélectionnez une vidéo",
        filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
    )
    if file_path:
        file_label.config(text=file_path)

def update_progress(process, progress_bar, on_complete):
    """
    Met à jour la barre de progression en analysant la sortie d'FFmpeg.
    :param process: Processus en cours d'exécution (FFmpeg).
    :param progress_bar: Widget de la barre de progression.
    :param on_complete: Callback à exécuter à la fin de la compression.
    :return: None
    """
    duration = None
    while True:
        output = process.stdout.readline()
        if output == "" and process.poll() is not None:
            break

        # Extraire la durée totale de la vidéo
        if "Duration" in output:
            match = re.search(r"Duration: (\d+):(\d+):(\d+)\.(\d+)", output)
            if match:
                hours, minutes, seconds, milliseconds = map(int, match.groups())
                duration = hours * 3600 + minutes * 60 + seconds + milliseconds / 1000

        # Extraire le temps traité et mettre à jour la barre de progression
        if "time=" in output and duration:
            match = re.search(r"time=(\d+):(\d+):(\d+)\.(\d+)", output)
            if match:
                hours, minutes, seconds, milliseconds = map(int, match.groups())
                current_time = hours * 3600 + minutes * 60 + seconds + milliseconds / 1000
                progress = (current_time / duration) * 100
                progress_bar["value"] = progress

    # Marquer la progression comme terminée
    progress_bar["value"] = 100
    on_complete()

def compress_video():
    """
    Lance la compression vidéo dans un thread séparé pour éviter de bloquer l'interface.
    :return: None
    """
    # Vérifier si un fichier a été sélectionné
    input_file = file_label.cget("text")
    if not input_file or input_file == "Aucun fichier sélectionné":
        messagebox.showerror("Erreur", "Veuillez sélectionner un fichier vidéo.")
        return

    # Définir les paramètres de compression
    output_file = input_file.rsplit(".", 1)[0] + "_compressed.mp4"
    ffmpeg_command = [
        "ffmpeg", "-i", input_file,
        "-vcodec", "libx264", "-crf", "23",
        "-y",  # Overwrite sans confirmation
        output_file
    ]

    def on_complete():
        """
        Callback exécuté une fois la compression terminée.
        """
        process_status.set("Compression terminée !")
        messagebox.showinfo("Succès", f"Compression terminée :\n{output_file}")

    def run_compression():
        """
        Exécute la commande FFmpeg et met à jour la progression.
        """
        try:
            process = subprocess.Popen(
                ffmpeg_command,
                stdout=subprocess.PIPE,
                stderr=subprocess.STDOUT,
                universal_newlines=True,
                bufsize=1
            )
            update_progress(process, progress_bar, on_complete)
        except Exception as e:
            messagebox.showerror("Erreur", f"Une erreur est survenue : {e}")

    # Mettre à jour le statut et lancer le thread
    process_status.set("Compression en cours...")
    threading.Thread(target=run_compression, daemon=True).start()

# Interface graphique
root = tk.Tk()
root.title("Compresseur Vidéo v1.1")

# Conteneur principal
frame = tk.Frame(root, padx=20, pady=20)
frame.pack()

# Label pour afficher le fichier sélectionné
file_label = tk.Label(frame, text="Aucun fichier sélectionné", wraplength=400)
file_label.pack(pady=10)

# Bouton pour sélectionner un fichier
select_button = tk.Button(frame, text="Sélectionner un fichier", command=select_file)
select_button.pack(pady=5)

# Bouton pour lancer la compression
compress_button = tk.Button(frame, text="Compresser la vidéo", command=compress_video)
compress_button.pack(pady=5)

# Barre de progression
progress_bar = ttk.Progressbar(frame, orient="horizontal", length=400, mode="determinate")
progress_bar.pack(pady=10)

# Label pour le statut du processus
process_status = tk.StringVar(value="En attente d'une action...")
status_label = tk.Label(frame, textvariable=process_status, wraplength=400)
status_label.pack(pady=5)

# Lancement de la boucle principale de l'interface
root.mainloop()
