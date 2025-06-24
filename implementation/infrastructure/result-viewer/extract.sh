#!/bin/bash

# This script recursively finds all .tar.gz files in the current directory,
# extracts them, moves the contents to the original archive's location,
# and then cleans up the archive and temporary extraction folder.

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# The name of the temporary directory for extractions.
BACKUP_DIR="backup-data"

# --- Main Logic ---

echo "Starting the extraction process..."

# Use find to locate all .tar.gz files from the current working directory (.).
# The -print0 and nullglob (`''`) with xargs handle filenames with spaces or special characters.
find . -type f -name "*.tar.gz" -print0 | while IFS= read -r -d '' archive; do
    
    # --- 1. Get File Paths and Names ---
    
    # Get the directory where the .tar.gz file is located.
    # e.g., if archive is "./data/my_archive.tar.gz", original_path is "./data"
    original_path=$(dirname "$archive")

    # Get the filename without the path.
    # e.g., if archive is "./data/my_archive.tar.gz", filename is "my_archive.tar.gz"
    filename=$(basename "$archive")

    # Get the base name of the file (without .tar.gz extension).
    # e.g., if filename is "my_archive.tar.gz", base_name is "my_archive"
    base_name="${filename%.tar.gz}"

    # Define the path for the temporary extraction directory.
    # e.g., "./backup-data/my_archive"
    extract_path="./$BACKUP_DIR/$base_name"

    echo "----------------------------------------------------"
    echo "Processing archive: $archive"

    # --- 2. Create Directory and Extract ---

    echo "Creating temporary extraction directory: $extract_path"
    # Create the nested directory structure. The -p flag ensures parent directories are created if they don't exist.
    mkdir -p "$extract_path"

    echo "Extracting '$archive' to '$extract_path'..."
    # Extract the tar.gz file into the newly created directory.
    # -x: extract
    # -z: decompress with gzip
    # -f: specify archive file
    # -C: change to directory before extracting
    tar -xzf "$archive" -C "$extract_path"

    # --- 3. Move Extracted Contents ---
    
    echo "Moving contents from '$extract_path' to '$original_path'..."
    # Move all files and folders (including hidden ones) from the extraction directory
    # to the original path where the .tar.gz file was located.
    # The `shopt -s dotglob` command makes the '*' glob include hidden files (dotfiles).
    (shopt -s dotglob; mv "$extract_path"/* "$original_path/")

    # --- 4. Cleanup ---
    
    echo "Cleaning up..."
    
    # Remove the original .tar.gz file.
    echo "Deleting archive: $archive"
    rm "$archive"
    
    # The temporary extraction directory should now be empty.
    # We can remove it and its parent if the parent is also empty.
    # rmdir only removes empty directories.
    echo "Deleting temporary directory: $extract_path"
    rmdir "$extract_path"

    # Attempt to remove the main backup-data directory.
    # This will only succeed if it's empty (i.e., this was the last archive processed).
    # The `|| true` prevents the script from exiting if the directory is not empty.
    rmdir "./$BACKUP_DIR" 2>/dev/null || true
    
    echo "Processing for '$archive' complete."
    echo "----------------------------------------------------"
    
done

echo "All .tar.gz files have been processed."

