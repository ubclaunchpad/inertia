const webpack = require('webpack');
const MinifyPlugin = require('babel-minify-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const mainConfig = require('./webpack.config');

const HtmlWebpackPluginConfig = new HtmlWebpackPlugin({
  template: './index.html',
  filename: 'index.html',
  inject: 'body',
});

const config = {
  mode: 'production',
  entry: ['babel-polyfill', './index.js'],
  output: {
    path: `${__dirname}/public/`,
    filename: 'bundle.js',
  },
  module: mainConfig.module,
  plugins: [
    new webpack.EnvironmentPlugin(['NODE_ENV']),
    HtmlWebpackPluginConfig,
    new MinifyPlugin(),
  ],
};

module.exports = config;
