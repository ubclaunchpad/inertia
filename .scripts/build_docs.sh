#! /bin/bash

set -e

###
# This script is a wonderful hack I made to get around the fact that Slate
# (https://github.com/lord/slate) documentation recommends that you fork the
# repository and then write your documentation inside it.
#
# To convert Slate into a pure "doc builder", this script does a few things:
#   * clone my (Robert's) fork of Slate into temporary /docs_build
#   * write and sed custom configuration into the repository
#   * symlink our doc assets (primarily in /docs_src, but also a few image assets
#     from .static) into the cloned Slate repository
#   * build the documentation into /docs, which is deployed by GH-pages
#
# The symlink + config hacks also allows live-reload local dev deployment (using
# the 'make run-docs' target in /Makefile).
###

# Get Slate
echo "[INFO] Getting Slate"
mkdir -p docs_build
cd docs_build
if [ ! -d slate ]; then
  # Using Robert's fork for now, for extra swag and configuration options not
  # available in vanilla Slate
  git clone https://github.com/bobheadxi/slate.git
else
  echo "[INFO] Slate already present in docs_build/slate"
fi

# Add custom config to Slate
echo "[INFO] Hacking Slate configuration"
TEMPLATE_FILES_WATCH="files.watch :source, path: File.join(root, '../../docs_src')"
if ! grep -q "$TEMPLATE_FILES_WATCH" slate/config.rb ; then
  # We want to symlink our doc assets into the repo, so to retain live reload
  # functionality, we want the generator Slate uses (Middleman) to watch our
  # symlink sources for changes too
  echo "$TEMPLATE_FILES_WATCH" \
    >> slate/config.rb
fi
if ! grep -q "<%= favicon_tag 'favicon.ico' %>" slate/source/layouts/layout.erb ; then
  # This inserts a favicon reference into the <head /> element of the
  # documentation layout
  sed -i '' '/<head>/a\
  <%= favicon_tag '\''favicon\.ico'\'' %>
  ' slate/source/layouts/layout.erb
fi

# Set up Slate for build
echo "[INFO] Linking assets"
ln -fs "$(dirname "$(pwd)")"/docs_src/index.html.md \
  slate/source/index.html.md
ln -fs "$(dirname "$(pwd)")"/docs_src/stylesheets/_variables.scss \
  slate/source/stylesheets/_variables.scss
ln -fs "$(dirname "$(pwd)")"/.static/inertia.png \
  slate/source/images/logo.png
ln -fs "$(dirname "$(pwd)")"/.static/favicon.ico \
  slate/source/images/favicon.ico
echo "[INFO] Installing Slate dependencies"
cd slate
bundle install

# Execute build
echo "[INFO] Building documentation"
rm -rf docs
bundle exec middleman build --clean --build-dir=../../docs
