import os
import re
import shutil
import tkinter as tk
from tkinter import filedialog

def rename_files(folder):
    # Remove any preexisting temporary folder
    temp_dir = os.path.join(folder, "temp_renamer")
    if os.path.exists(temp_dir):
        shutil.rmtree(temp_dir)
    os.makedirs(temp_dir)
    
    # Updated pattern to also match numbers in filenames ending with _compressed
    base_pattern = re.compile(r'^(\d{3})(\.[^\.]+)$')
    compressed_pattern = re.compile(r'^(\d{3})_compressed(\.[^\.]+)$')
    
    max_num = -1
    to_rename = []
    correct_files = []
    
    for f in os.listdir(folder):
        full_path = os.path.join(folder, f)
        if os.path.isfile(full_path):
            # Check for regular numbered files
            m = base_pattern.match(f)
            if m:
                num = int(m.group(1))
                max_num = max(max_num, num)
                correct_files.append(f)
                continue
            
            # Check for _compressed files to update max_num but don't move them
            m = compressed_pattern.match(f)
            if m:
                num = int(m.group(1))
                max_num = max(max_num, num)
                continue
                
            # If file doesn't match either pattern and isn't a _compressed file
            if '_compressed' not in f:
                to_rename.append(f)

    # Move correct files to temporary folder
    for f in correct_files:
        shutil.move(os.path.join(folder, f), os.path.join(temp_dir, f))

    # Rename files starting from the next available number
    next_num = max_num + 1 if max_num != -1 else 0
    for f in to_rename:
        src_path = os.path.join(folder, f)
        if os.path.isfile(src_path):
            _, ext = os.path.splitext(f)
            new_name = f"{next_num:03}{ext}"
            os.rename(src_path, os.path.join(folder, new_name))
            next_num += 1

    # Move back the original numbered files
    for f in os.listdir(temp_dir):
        shutil.move(os.path.join(temp_dir, f), os.path.join(folder, f))
    os.rmdir(temp_dir)
    print("Renaming complete.")

def main():
    # Select folder
    root = tk.Tk()
    root.withdraw()
    folder = filedialog.askdirectory(title="Select Folder to Rename Files")
    if not folder:
        print("No folder selected.")
        return
    rename_files(folder)

if __name__ == "__main__":
    main()
