// Imports
const request = require('request');
const path = require('path');
const mkdirp = require('mkdirp');
const fs = require('fs');
const exec = require('child_process').exec;
const pkg = require('../package.json');

// Mappings
const ARCH_PLATFORM_MAPPINGS = {
  ia32: '386',
  x64: 'amd64',
  arm: 'arm',
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows',
  freebsd: 'freebsd'
};

// Inertia options
const opts = {
  binPath: './bin',
  binName: 'inertia',
  binversion: pkg.version
};

const exe = process.platform === 'win32' ? '.exe' : '';

// Builds Inertia download url depending on user machine
function getBinURL () {
  // Verify system
  if (!(process.arch in ARCH_PLATFORM_MAPPINGS) || !(process.platform in ARCH_PLATFORM_MAPPINGS)) {
    console.error(`Inertia CLI is not supported on this system: ${process.platform}-${process.arch}`);
    return;
  }

  // Build download Url
  return `https://github.com/ubclaunchpad/inertia/releases/download/v${opts.binversion}/inertia.v${opts.binversion}.${ARCH_PLATFORM_MAPPINGS[process.platform]}.${ARCH_PLATFORM_MAPPINGS[process.arch]}${exe}`;
}

// Returns the location of bin, some users might change this
function getNpmBinLocation (callback) {
  exec('npm bin', (err, stdout, stderr) => {
    let dir;
    if (err || stderr || !stdout || stdout.length === 0) {
      // Infer path from enviroment variables on err
      const env = process.env;
      if (env && env.npm_config_prefix) {
        dir = path.join(env.npm_config_prefix, 'bin');
      }
    } else {
      dir = stdout.trim();
    }

    mkdirp.sync(dir);
    callback(null, dir);
  });
}

// Copy inertia bin file to final location
function copyToFinalLocation (binName, binPath, callback) {
  if (!fs.existsSync(path.join(binPath, binName))) return callback(new Error('Couldn\'t find binary file'));
  getNpmBinLocation((err, installationPath) => {
    if (err) return callback(new Error('Error getting binary installation path from `npm bin`'));
    // Move the binary file to final location
    fs.renameSync(path.join(binPath, binName), path.join(installationPath, binName + exe));
    // Set file permissions
    fs.chmodSync(path.join(installationPath, binName + exe), 755);
    console.log('Run `inertia --version` to check if the installation was successful!\n');
    callback(null);
  });
}

// Wrapper that handdles the complete installation proccess
module.exports.install = function (callback) {
  mkdirp.sync(opts.binPath);

  const req = request({
    uri: getBinURL()
  });

  req.on('error', callback.bind(null, `Error downloading from URL: ${opts.url}`));
  req.on('response', (res) => {
    if (res.statusCode !== 200) return callback(new Error('Error downloading binary'));

    // create file write stream
    const writeStream = fs.createWriteStream(path.resolve(opts.binPath, opts.binName));
    // setup piping
    res.pipe(writeStream);
    writeStream.on('finish', copyToFinalLocation.bind(null, opts.binName, opts.binPath, callback));
    writeStream.on('error', callback.bind(null, 'Error downloading binary'));
  });
};

// Wrapper that handdles the complete un-installation proccess
module.exports.uninstall = function (callback) {
  getNpmBinLocation((err, installationPath) => {
    if (err) callback(new Error('Error finding binary installation directory'));

    let fullPath = path.join(installationPath, opts.binName + exe);
    if (fs.existsSync(fullPath)) { // Uinstall only if present
      try {
        fs.unlinkSync(fullPath);
      } catch (err) {
        console.log('Error while uninstalling');
      }
    }
    return callback(null);
  });
};
