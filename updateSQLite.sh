#!/bin/bash
set -e

if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <zip_url>"
  exit 1
fi

ZIP_URL="$1"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TMP_DIR="$SCRIPT_DIR/tmp"
mkdir -p "$TMP_DIR"
ZIP_FILE="$TMP_DIR/archive.zip"

# Télécharger le fichier zip
curl -L "$ZIP_URL" -o "$ZIP_FILE"

# Extraire dans le dossier temporaire
unzip -j "$ZIP_FILE" -d "$TMP_DIR"

# Copier les deux fichiers à la racine du script
cp "$TMP_DIR/sqlite3.c" "$SCRIPT_DIR/sqlite3-binding.c"
cp "$TMP_DIR/sqlite3.h" "$SCRIPT_DIR/sqlite3-binding.h"
cp "$TMP_DIR/sqlite3ext.h" "$SCRIPT_DIR/sqlite3ext.h"

# Nettoyer
rm -rf "$TMP_DIR"

echo "Fichiers mis à jour avec succès !"