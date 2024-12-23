import unittest
from queue import Queue

class TestQueueCompression(unittest.TestCase):
    """
    Classe de tests pour vérifier le traitement d'une file d'attente de vidéos à compresser.
    """

    def test_queue_processing(self):
        """
        Vérifie que la file d'attente est correctement traitée.
        """
        # Création de la file d'attente et ajout de fichiers simulés
        queue = Queue()
        queue.put("video1.mp4")
        queue.put("video2.mp4")

        # Vérifie que deux éléments ont été ajoutés à la file
        self.assertEqual(queue.qsize(), 2)

        # Simule le traitement des fichiers dans la file
        while not queue.empty():
            file = queue.get()
            # Simulation de la compression
            print(f"Compression de : {file}")
            queue.task_done()

        # Vérifie que la file est vide après le traitement
        self.assertEqual(queue.qsize(), 0)

if __name__ == "__main__":
    unittest.main()

"""
Explication des tests pour le fichier : test_queue_compression.py

1. Test du traitement de la file d'attente :
   - Ajoute deux fichiers fictifs ("video1.mp4" et "video2.mp4") dans une file d'attente.
   - Vérifie que le nombre d'éléments dans la file correspond à l'attendu (2).
   - Simule le traitement en affichant un message pour chaque fichier extrait de la file.
   - Vérifie qu'après le traitement, la file est vide.

Ce test est conçu pour s'assurer que la file d'attente est manipulée correctement
et que tous les fichiers ajoutés sont traités sans erreurs.
"""
