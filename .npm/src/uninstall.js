#!/usr/bin/env node


const util = require('./utils.js');

util.uninstall((err) => {
  if (err) {
    console.error(err);
    process.exit(1);
  } else {
    process.exit(0);
  }
});
