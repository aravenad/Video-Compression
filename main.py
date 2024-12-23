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

def update_progress(process, progress_widget):
    """
    Met à jour la barre de progression en analysant la sortie d'FFmpeg.
    :param process: Processus en cours d'exécution (FFmpeg).
    :param progress_widget: Widget de la barre de progression.
    :return: None
    """
    duration = None  # Durée totale de la vidéo en secondes
    while True:
        # Lire la sortie d'FFmpeg ligne par ligne
        output = process.stdout.readline()
        # Sortir de la boucle si le processus est terminé
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
                progress = (current_time / duration) * 100  # Calcul du pourcentage
                progress_widget["value"] = progress  # Mise à jour de la barre

    # Mettre la barre à 100% une fois terminé
    progress_widget["value"] = 100

def compress_video():
    """
    Compresse la vidéo sélectionnée en utilisant FFmpeg.
    Affiche une barre de progression pour indiquer l'état de la compression.
    :return: None
    """
    # Vérifier si un fichier a été sélectionné
    input_file = file_label.cget("text")
    if not input_file or input_file == "Aucun fichier sélectionné":
        messagebox.showerror("Erreur", "Veuillez sélectionner un fichier vidéo.")
        return

    # Définir les paramètres de compression par défaut
    output_file = input_file.rsplit(".", 1)[0] + "_compressed.mp4"
    ffmpeg_command = [
        "ffmpeg", "-i", input_file,
        "-vcodec", "libx264", "-crf", "23",
        "-y",  # Overwrite sans confirmation
        output_file
    ]

    try:
        # Réinitialiser la barre de progression
        progress_bar["value"] = 0

        # Lancer FFmpeg avec redirection de sortie
        process = subprocess.Popen(
            ffmpeg_command,
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,  # Combiner stdout et stderr
            universal_newlines=True,
            bufsize=1
        )

        # Lancer un thread pour traiter la progression sans bloquer l'interface
        threading.Thread(target=update_progress, args=(process, progress_bar), daemon=True).start()

        # Attendre la fin du processus
        process.wait()
        if process.returncode == 0:
            messagebox.showinfo("Succès", f"Compression terminée :\n{output_file}")
        else:
            messagebox.showerror("Erreur", "La compression a échoué.")
    except Exception as e:
        # Afficher une erreur générique en cas de problème
        messagebox.showerror("Erreur", f"Une erreur est survenue : {e}")

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

# Lancement de la boucle principale de l'interface
root.mainloop()
