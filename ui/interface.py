import tkinter as tk
from tkinter import filedialog, ttk
from queue import Queue
import threading
from core.compression import compress_video

# File d'attente globale
compression_queue = Queue()

def start_app():
    """
    Initialise et lance l'interface utilisateur.
    """
    root = tk.Tk()
    root.title("Compresseur Vidéo")

    # Conteneur principal
    frame = tk.Frame(root, padx=20, pady=20)
    frame.pack()

    # Label pour afficher les fichiers sélectionnés
    file_label = tk.Label(frame, text="Aucun fichier sélectionné", wraplength=400)
    file_label.pack(pady=10)

    def select_file():
        file_path = filedialog.askopenfilename(
            title="Sélectionnez une vidéo",
            filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
        )
        if file_path:
            file_label.config(text=file_path)

    def select_files():
        file_paths = filedialog.askopenfilenames(
            title="Sélectionnez des vidéos",
            filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
        )
        if file_paths:
            for file_path in file_paths:
                compression_queue.put(file_path)
            process_status.set(f"{len(file_paths)} fichiers ajoutés à la file.")

    def start_compression():
        """
        Traite la file d'attente de compression.
        """
        def worker():
            while not compression_queue.empty():
                file = compression_queue.get()
                compress_video(file, progress_bar, process_status)
                compression_queue.task_done()

        threading.Thread(target=worker, daemon=True).start()

    # Bouton pour sélectionner un fichier
    select_button = tk.Button(frame, text="Sélectionner un fichier", command=select_file)
    select_button.pack(pady=5)

    # Bouton pour sélectionner plusieurs fichiers
    select_files_button = tk.Button(frame, text="Sélectionner des fichiers", command=select_files)
    select_files_button.pack(pady=5)

    # Bouton pour lancer la compression des fichiers dans la file d'attente
    compress_all_button = tk.Button(frame, text="Compresser les vidéos", command=start_compression)
    compress_all_button.pack(pady=5)

    # Barre de progression
    progress_bar = ttk.Progressbar(frame, orient="horizontal", length=400, mode="determinate")
    progress_bar.pack(pady=10)

    # Label pour le statut du processus
    process_status = tk.StringVar(value="En attente d'une action...")
    status_label = tk.Label(frame, textvariable=process_status, wraplength=400)
    status_label.pack(pady=5)

    # Lancement de l'interface graphique
    root.mainloop()
