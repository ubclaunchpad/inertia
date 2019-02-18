#!/usr/bin/env node

const fs = require('fs-extra');

fs.copy('../../README.md', 'README.md', (err) => {
  if (err) throw err;
});

fs.copy('../../.static', '.static', (err) => {
  if (err) throw err;
});
