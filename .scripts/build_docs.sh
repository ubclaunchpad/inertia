#! /bin/bash

set -e

# Get Slate
echo "[INFO] Getting Slate"
mkdir -p docs_build
cd docs_build
if [ ! -d slate ]; then
  git clone https://github.com/lord/slate.git
else
  echo "[INFO] Slate already present in docs_build/slate"
fi

# Set up Slate for build
echo "[INFO] Linking assets"
ln -fs "$(dirname "$(pwd)")"/docs_src/index.html.md \
  slate/source/index.html.md
ln -fs "$(dirname "$(pwd)")"/docs_src/_variables.scss \
  slate/source/stylesheets/_variables.scss
ln -fs "$(dirname "$(pwd)")"/.static/inertia-with-name.png \
  slate/source/images/logo.png
echo "[INFO] Installing Slate dependencies"
cd slate
bundle install

# Execute build
echo "[INFO] Building documentation"
bundle exec middleman build --clean

# Move output to /docs
echo "[INFO] Migrating build to /docs"
cd ../../
rm -rf docs
mkdir -p docs
mv -v docs_build/slate/build/* docs
