import tkinter as tk
from tkinter import filedialog, ttk
from core.compression import compress_video

def start_app():
    """
    Initialise et lance l'interface utilisateur.
    """
    root = tk.Tk()
    root.title("Compresseur Vidéo")

    # Conteneur principal
    frame = tk.Frame(root, padx=20, pady=20)
    frame.pack()

    # Label pour afficher le fichier sélectionné
    file_label = tk.Label(frame, text="Aucun fichier sélectionné", wraplength=400)
    file_label.pack(pady=10)

    def select_file():
        file_path = filedialog.askopenfilename(
            title="Sélectionnez une vidéo",
            filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
        )
        if file_path:
            file_label.config(text=file_path)

    # Bouton pour sélectionner un fichier
    select_button = tk.Button(frame, text="Sélectionner un fichier", command=select_file)
    select_button.pack(pady=5)

    # Barre de progression
    progress_bar = ttk.Progressbar(frame, orient="horizontal", length=400, mode="determinate")
    progress_bar.pack(pady=10)

    # Label pour le statut du processus
    process_status = tk.StringVar(value="En attente d'une action...")
    status_label = tk.Label(frame, textvariable=process_status, wraplength=400)
    status_label.pack(pady=5)

    # Bouton pour lancer la compression
    compress_button = tk.Button(
        frame,
        text="Compresser la vidéo",
        command=lambda: compress_video(file_label.cget("text"), progress_bar, process_status)
    )
    compress_button.pack(pady=5)

    # Lancement de l'interface graphique
    root.mainloop()
