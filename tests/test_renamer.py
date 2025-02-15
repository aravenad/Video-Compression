import os
import sys
# Add project root to sys.path
sys.path.append(os.path.join(os.path.dirname(__file__), '..'))
import shutil
import tempfile
import pytest
from renamer import rename_files

@pytest.fixture
def temp_folder():
    folder = tempfile.mkdtemp()
    yield folder
    shutil.rmtree(folder)

def test_rename_files_with_no_existing_renamed(temp_folder):
    # Create files in random order
    filenames = ['video.mp4', 'clip.avi', 'movie.mkv']
    for name in filenames:
        open(os.path.join(temp_folder, name), 'a').close()
    rename_files(temp_folder)
    # Check that files are renamed incrementally from 000
    expected = {f"{i:03}" for i in range(len(filenames))}
    result = {os.path.splitext(f)[0] for f in os.listdir(temp_folder)}
    assert expected == result

def test_rename_files_preserves_existing_format(temp_folder):
    # Simulate existing correctly renamed file and other files
    existing = "012.mp4"
    open(os.path.join(temp_folder, existing), 'a').close()
    other_files = ['random.avi', 'test.mkv']
    for name in other_files:
        open(os.path.join(temp_folder, name), 'a').close()
    rename_files(temp_folder)
    # The existing file should remain untouched
    assert os.path.exists(os.path.join(temp_folder, existing))
    # Other files should continue from 013 onward
    expected_nums = {"012", "013", "014"}
    result = {os.path.splitext(f)[0] for f in os.listdir(temp_folder)}
    assert expected_nums == result
