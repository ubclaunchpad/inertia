#! /bin/bash

set -e

# Get Slate
echo "[INFO] Getting Slate"
mkdir -p docs_build
cd docs_build
if [ ! -d slate ]; then
  git clone https://github.com/lord/slate.git
fi

# Set up Slate for build
echo "[INFO] Installing Slate dependencies"
cd slate
bundle install
echo "[INFO] Linking assets"
ln -f ../../docs_src/index.html.md \
  source/index.html.md
ln -f ../../docs_src/_variables.scss \
  source/stylesheets/_variables.scss
ln -f ../../.static/inertia-with-name.png \
  source/images/logo.png

# Execute build
echo "[INFO] Building documentation"
bundle exec middleman build --clean

# Move output to /docs
echo "[INFO] Migrating build to /docs"
cd ../../
rm -rf docs
mkdir -p docs
mv -v docs_build/slate/build/* docs
