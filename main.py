import tkinter as tk
from tkinter import filedialog, messagebox
import subprocess

def select_file():
    """
    Ouvre une boîte de dialogue pour sélectionner un fichier vidéo.
    :return: None
    """

    file_path = filedialog.askopenfilename(
        title="Sélectionnez une vidéo",
        filetypes=[("Fichiers vidéo", "*.mp4 *.mkv *.avi *.mov *.flv *.wmv")]
    )
    # Mettre à jour le label avec le chemin du fichier sélectionné
    if file_path:
        file_label.config(text=file_path)

def compress_video():
    """
    Compresse la vidéo sélectionnée en utilisant FFmpeg.
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
        output_file
    ]

    # Exécuter la commande FFmpeg
    try:
        subprocess.run(ffmpeg_command, check=True)
        messagebox.showinfo("Succès", f"Compression terminée :\n{output_file}")
    except subprocess.CalledProcessError:
        messagebox.showerror("Erreur", "La compression a échoué.")

# Interface graphique
root = tk.Tk()
root.title("Compresseur Vidéo v1.0")

frame = tk.Frame(root, padx=20, pady=20)
frame.pack()

file_label = tk.Label(frame, text="Aucun fichier sélectionné", wraplength=400)
file_label.pack(pady=10)

select_button = tk.Button(frame, text="Sélectionner un fichier", command=select_file)
select_button.pack(pady=5)

compress_button = tk.Button(frame, text="Compresser la vidéo", command=compress_video)
compress_button.pack(pady=5)

root.mainloop()
