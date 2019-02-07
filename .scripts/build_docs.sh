#! /bin/bash

# Get Slate
mkdir -p docs_build
cd docs_build || exit
git clone https://github.com/lord/slate.git

# Set up Slate for build
cd slate || exit
bundle install
cp -f ../../docs_src/index.html.md \
  source/index.html.md
cp -f ../../docs_src/_variables.scss \
  source/stylesheets/_variables.scss
cp -f ../../.static/inertia-with-name.png \
  source/images/logo.png

# Execute build
bundle exec middleman build --clean

# Move output to /docs
cd ../../ || exit
rm -rf docs
mkdir -p docs
mv -v docs_build/slate/build/* docs
