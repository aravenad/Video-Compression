import shutil
import os
import logging
from queue import Queue
import configparser

# Load configuration from config.ini
config = configparser.ConfigParser()
config.read("config.ini")

compression_queue = Queue()
MAX_QUEUE_SIZE = 10
active_compressions = []
current_output_file = None

compression_settings = {
    'quality': int(config.get("Compression", "quality", fallback="23")),
    'preset': config.get("Compression", "preset", fallback="medium"),  # Fixed: removed extra parenthesis
    'threads': int(config.get("Compression", "threads", fallback="2")),
    'use_nvenc': config.getboolean("Compression", "use_nvenc", fallback=False)
}

# New: List to track pending files
pending_files = []

# List to track successfully compressed original files for deletion later
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
    else:
        ffmpeg = shutil.which('ffmpeg')
    if not ffmpeg:
        raise RuntimeError("ffmpeg non trouvé. Veuillez l'installer.")
    return ffmpeg

def create_ffmpeg_command(input_file, output_file):
    """Create ffmpeg command with current settings"""
    base_cmd = [
        get_ffmpeg_path(),
        "-i", input_file,
        "-threads", str(compression_settings.get('threads', 2))
    ]
    if compression_settings.get('use_nvenc'):
        base_cmd += ["-vcodec", "h264_nvenc"]
    else:
        base_cmd += ["-vcodec", "libx264"]
    base_cmd += [
        "-crf", str(compression_settings['quality']),
        "-preset", compression_settings['preset'],
        "-y",
        "-progress", "pipe:1",
        output_file
    ]
    return base_cmd

def format_size(size):
    """
    Format size to human readable string using French units.
    """
    units = [' o', ' Ko', ' Mo', ' Go']
    i = 0
    while size >= 1024 and i < len(units) - 1:
        size /= 1024
        i += 1
    return f"{size:.1f}{units[i]}"

def estimate_output_size(input_size, quality):
    """
    Estimate output file size based on input size, quality and encoder.
    Uses a compression ratio that varies linearly from 2% (quality=10)
    to 10% (quality=51) of the original size.
    Applies a correction factor which is increased for GPU (NVENC) acceleration.
    """
    quality = max(quality, 10)
    ratio = ((quality - 10) / 41) * 0.08 + 0.02
    correction = 1.17 if compression_settings.get('use_nvenc') else 1.15
    estimated_size = input_size * ratio * correction
    return max(estimated_size, input_size * 0.02 * correction)
