import shutil
import os
import logging
from queue import Queue

compression_queue = Queue()
MAX_QUEUE_SIZE = 10
active_compressions = []
current_output_file = None

compression_settings = {
    'quality': 23,
    'preset': 'medium',
    'threads': 2  # New: limit number of CPU threads used by ffmpeg
}

# New: list to track successfully compressed original files for deletion later
compressed_files = []

# UI elements
est_size_label = None
progress_bar = None
status_label = None

def cleanup_current_file():
    """Clean up the current output file if it exists"""
    global current_output_file
    if current_output_file and os.path.exists(current_output_file):
        try:
            os.remove(current_output_file)
            logging.info(f"Deleted incomplete file: {current_output_file}")
        except OSError as e:
            logging.error(f"Could not delete file {current_output_file}: {e}")
        current_output_file = None

def get_ffmpeg_path():
    """Get ffmpeg executable path"""
    if os.name == 'nt':  # Windows
        ffmpeg = shutil.which('ffmpeg.exe')
    if not ffmpeg:
        raise RuntimeError("ffmpeg non trouvé. Veuillez l'installer.")
    return ffmpeg

def create_ffmpeg_command(input_file, output_file):
    """Create ffmpeg command with current settings"""
    return [
        get_ffmpeg_path(),
        "-i", input_file,
        "-vcodec", "libx264",
        "-threads", str(compression_settings.get('threads', 2)),  # New: set thread count
        "-crf", str(compression_settings['quality']),
        "-preset", compression_settings['preset'],
        "-y",
        "-progress", "pipe:1",
        output_file
    ]

def format_size(size):
    """
    Format size to human readable string using French units.
    """
    units = ['o', 'Ko', 'Mo', 'Go']
    i = 0
    while size >= 1024 and i < len(units) - 1:
        size /= 1024
        i += 1
    return f"{size:.1f}{units[i]}"

def estimate_output_size(input_size, quality):
    """
    Estimate output file size based on input size and quality.
    Uses a compression ratio that varies linearly from 2% (quality=10)
    to 10% (quality=51) of the original size and applies a correction factor of 1.15.
    """
    quality = max(quality, 10)
    ratio = ((quality - 10) / 41) * 0.08 + 0.02
    estimated_size = input_size * ratio * 1.15  # Correction factor added
    return max(estimated_size, input_size * 0.02 * 1.15)

def update_size_estimation(input_file):
    """Update size estimation label"""
    global est_size_label
    if not est_size_label:
        return
        
    try:
        input_size = os.path.getsize(input_file)
        est_size = estimate_output_size(input_size, compression_settings['quality'])
        est_size_label.config(text=f"Taille estimée: {format_size(est_size)}")
    except Exception as e:
        logging.error(f"Error updating size estimation: {e}")
