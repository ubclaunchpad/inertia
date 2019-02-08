#! /bin/bash

set -e

# Get Slate
echo "[INFO] Getting Slate"
mkdir -p docs_build
cd docs_build
if [ ! -d slate ]; then
  # Using Robert's fork for now, for extra swag
  git clone https://github.com/bobheadxi/slate.git
else
  echo "[INFO] Slate already present in docs_build/slate"
fi

# Add custom config to Slate
echo "[INFO] Hacking Slate configuration"
echo "files.watch :source, path: File.join(root, '../../docs_src')" \
  >> slate/config.rb

# Set up Slate for build
echo "[INFO] Linking assets"
ln -fs "$(dirname "$(pwd)")"/docs_src/index.html.md \
  slate/source/index.html.md
ln -fs "$(dirname "$(pwd)")"/docs_src/stylesheets/_variables.scss \
  slate/source/stylesheets/_variables.scss
ln -fs "$(dirname "$(pwd)")"/.static/inertia.png \
  slate/source/images/logo.png
echo "[INFO] Installing Slate dependencies"
cd slate
bundle install

# Execute build
echo "[INFO] Building documentation"
rm -rf docs
bundle exec middleman build --clean --build-dir=../../docs
