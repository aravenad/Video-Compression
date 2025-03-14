# Video-Compression

## Description
**Video-Compression** est un logiciel Python basé sur FFmpeg pour compresser facilement des vidéos. Avec une interface intuitive, il offre des options de compression optimisées, une gestion par file d'attente et la possibilité de libérer de l'espace disque en supprimant les fichiers d'origine.
  
---
  
## Fonctionnalités
- **Exploration vidéo :** Sélectionnez un dossier et affichez uniquement les fichiers vidéo compatibles.
- **Compression personnalisée :** Utilisez des paramètres par défaut optimisés ou configurez vos propres options.
- **File d'attente :** Gérez plusieurs vidéos et traitez-les séquentiellement.
- **Gestion des fichiers :** Supprimez automatiquement les fichiers d'origine pour économiser de l'espace.
- **Interface conviviale :** Accessible aux utilisateurs novices et avancés.
  
---
  
## Technologies utilisées
- **Python** : Langage de développement principal.
- **FFmpeg** : Bibliothèque pour le traitement et la compression vidéo.
- **Tkinter** : Frameworks possibles pour l’interface graphique (à définir).
- **Subprocess / ffmpeg-python** : Outils pour l'intégration d'FFmpeg.
  
---
  
## Installation
1. Clonez ce dépôt :  
   git clone https://github.com/aravenad/Video-Compression.git  
   cd Video-Compression
2. Installez les dépendances nécessaires :  
   pip install -r requirements.txt
3. Assurez-vous que Ffmpeg est installé sur votre machine :  
   ffmpeg -version  
   *(Télécharger Ffmpeg ici : https://ffmpeg.org/download.html)*
  
---
  
## Utilisation
1. Exécutez l'application :
   python ./main.py
2. Sélectionnez un dossier contenant vos vidéos.
3. Ajoutez les vidéos à la file d'attente.
4. Configurez vos options de compression ou utilisez les paramètres par défaut.
5. Lancez la compression et profitez d'un processus automatisé !
  
---
  
## Licence
Ce projet est sous licence MIT : consultez le fichier LICENSE pour plus d'informations.
