#!/usr/bin/env node

'use strict';

const uninstall = require("./utils.js").uninstall;

uninstall(function(err){
    if(err){
        console.error(err);
        process.exit(1);
    }else{
        process.exit(0);
    }
});
