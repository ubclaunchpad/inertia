#!/usr/bin/env node

'use strict';

const install = require("./utils.js").install;

install(function(err){
    if(err){
        console.error(err);
        process.exit(1);
    }else{
        process.exit(0);
    }
});
