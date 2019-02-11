#!/usr/bin/env node


const utils = require('./utils.js')

utils.install((err) => {
  if (err) {
    console.error(err);
    process.exit(1);
  } else {
    process.exit(0);
  }
});
