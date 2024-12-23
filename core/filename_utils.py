import os

def generate_unique_filename(input_file):
    """
    Génère un nom de fichier unique en vérifiant l'existence de fichiers existants.
    """
    base_name, ext = os.path.splitext(input_file)
    output_file = f"{base_name}_compressed{ext}"
    counter = 1

    while os.path.exists(output_file):
        output_file = f"{base_name}_compressed-{counter}{ext}"
        counter += 1

    return output_file
