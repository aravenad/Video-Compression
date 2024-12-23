import unittest
import os
from core.filename_utils import generate_unique_filename

class TestFilenameUtils(unittest.TestCase):

    def setUp(self):
        """
        Prépare un environnement temporaire pour les tests.
        """
        self.temp_dir = "temp_test_dir"
        os.makedirs(self.temp_dir, exist_ok=True)
        self.temp_file = os.path.join(self.temp_dir, "video.mp4")
        with open(self.temp_file, "w") as f:
            f.write("")  # Crée un fichier temporaire

    def tearDown(self):
        """
        Nettoie l'environnement temporaire après les tests.
        """
        for file in os.listdir(self.temp_dir):
            os.remove(os.path.join(self.temp_dir, file))
        os.rmdir(self.temp_dir)

    def test_unique_filename_no_conflict(self):
        """
        Vérifie que le nom de fichier généré est correct si aucun conflit n'existe.
        """
        unique_name = generate_unique_filename(self.temp_file)
        expected_name = os.path.join(self.temp_dir, "video_compressed.mp4")
        self.assertEqual(unique_name, expected_name)

    def test_unique_filename_with_conflict(self):
        """
        Vérifie que le suffixe incrémental est ajouté en cas de conflit.
        """
        # Crée un fichier avec le nom attendu
        conflict_file = os.path.join(self.temp_dir, "video_compressed.mp4")
        with open(conflict_file, "w") as f:
            f.write("")

        unique_name = generate_unique_filename(self.temp_file)
        expected_name = os.path.join(self.temp_dir, "video_compressed-1.mp4")
        self.assertEqual(unique_name, expected_name)

if __name__ == "__main__":
    unittest.main()

"""
Explication des tests pour le fichier : test_filename_utils.py

Utilisation de setUp et tearDown :
- `setUp` prépare un répertoire temporaire pour simuler un environnement de fichiers sans impacter le système principal.
- `tearDown` supprime tous les fichiers et répertoires temporaires après chaque test pour garantir un état propre.

Tests inclus :
1. Test sans conflit de noms :
   - Vérifie que `generate_unique_filename` retourne un nom de fichier avec le suffixe `_compressed` lorsque le fichier d'origine n'a pas de conflit.

2. Test avec conflit de noms :
   - Simule l'existence d'un fichier déjà compressé.
   - Vérifie que la fonction incrémente correctement le suffixe numérique (`_compressed-1`, `_compressed-2`, etc.) jusqu'à trouver un nom unique.
"""
