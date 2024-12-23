import unittest
import subprocess
from unittest.mock import patch, MagicMock
from io import StringIO
from threading import Event
from core.compression import compress_video

# NOTE : Tester la compression réelle peut être long et dépendant du système, donc on utilise des tests simulés (mock).
class TestCompression(unittest.TestCase):

    @patch("core.compression.subprocess.Popen")
    def test_compression_execution(self, mock_popen):
        """
        Vérifie que la commande FFmpeg est appelée avec les bons arguments et que
        la progression atteint 100%.
        """
        # Mock du processus
        mock_process = MagicMock()
        mock_popen.return_value = mock_process
        mock_process.stdout = StringIO(
            "Duration: 00:00:10.000\n"
            "out_time_us=2000000\n"
            "out_time_us=5000000\n"
            "out_time_us=8000000\n"
            "out_time_us=10000000\n"
        )

        # Mock de la barre de progression (simule un accès direct via un dictionnaire)
        mock_progress_bar = {"value": 0}
        mock_status_label = MagicMock()

        # Utiliser un événement pour attendre la fin du thread
        thread_done = Event()

        def on_complete():
            """Callback appelé à la fin de la compression."""
            thread_done.set()

        input_file = "video.mp4"

        # Mock de la génération des noms uniques
        with patch("core.filename_utils.generate_unique_filename", return_value="video_compressed.mp4"):
            compress_video(input_file, mock_progress_bar, mock_status_label)

        # Attendre la fin du thread
        thread_done.wait(timeout=5)

        # Vérifie que subprocess.Popen est appelé avec les bons arguments
        mock_popen.assert_called_once_with(
            [
                "ffmpeg", "-i", "video.mp4", "-vcodec", "libx264", "-crf", "23",
                "-y", "-progress", "pipe:1", "video_compressed.mp4"
            ],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            universal_newlines=True,
            bufsize=1
        )

        # Vérifie que la progression atteint 100%
        self.assertEqual(mock_progress_bar["value"], 100, "La barre de progression n'a pas atteint 100%.")

        # Vérifie que le label de statut est correctement mis à jour
        mock_status_label.set.assert_any_call("Compression en cours...")
        mock_status_label.set.assert_any_call("Compression terminée.")

    @patch("core.compression.subprocess.Popen")
    def test_compression_invalid_input(self, mock_popen):
        """
        Vérifie qu'une exception est levée si l'entrée est invalide.
        """
        mock_progress_bar = {"value": 0}
        mock_status_label = MagicMock()

        with self.assertRaises(ValueError):
            compress_video(None, mock_progress_bar, mock_status_label)

        # Vérifie que subprocess.Popen n'est jamais appelé
        mock_popen.assert_not_called()

if __name__ == "__main__":
    unittest.main()

"""
Explication des tests pour le fichier : test_compression.py

Utilisation des mocks :
- Simule l’exécution de subprocess.Popen pour éviter d’exécuter réellement FFmpeg.
- Mock de stdout avec io.StringIO pour simuler un flux de sortie lisible par readline().
- Vérifie que la commande FFmpeg construite correspond exactement à celle attendue.

Tests inclus :
1. Test de la commande FFmpeg :
   - Vérifie que subprocess.Popen est appelé avec les bons arguments, y compris les options comme `-progress pipe:1`.
   - Simule les logs de FFmpeg dans stdout pour tester la progression.
   - Vérifie que la barre de progression atteint `100%`.
   - Vérifie que le statut est mis à jour avec "Compression en cours..." et "Compression terminée."

2. Test d'entrée invalide :
   - Vérifie que la fonction `compress_video` lève une exception (ValueError) si aucun fichier valide n'est fourni.
   - Assure que subprocess.Popen n'est pas appelé lorsque l'entrée est invalide.
"""
