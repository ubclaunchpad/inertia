const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const config = {
  mode: 'development',
  entry: ['babel-polyfill', './index.js'],
  output: {
    path: `${__dirname}/public/`,
    filename: 'bundle.js',
  },
  devServer: {
    port: 7900,
    inline: true,
    contentBase: './public',
    publicPath: '/web/',
    proxy: {
      '/': { target: 'https://127.0.0.1:4303', secure: false },
    },
  },
  devtool: 'inline-source-map',
  module: {
    rules: [
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        use: [
          {
            loader: 'babel-loader',
            options: {
              presets: ['es2015', 'stage-3', 'react'],
            },
          },
        ],
      },
      {
        test: /\.sass($|\?)/,
        exclude: /node_modules/,
        use: ['style-loader', 'css-loader', 'sass-loader'],
        resolve: {
          mainFiles: ['index.sass'],
        },
      },
      {
        test: /\.png($|\?)|\.jpg($|\?)|\.ico($|\?)|\.woff($|\?)|\.woff2($|\?)|\.ttf($|\?)|\.eot($|\?)/,
        loader: 'url-loader',
      },
      {
        test: /\.svg$/,
        use: ['babel-loader', 'react-svg-loader'],
      },
    ],
  },
  plugins: [
    new webpack.EnvironmentPlugin(['NODE_ENV']),
    new webpack.EnvironmentPlugin(['INERTIA_API']),
    new webpack.DefinePlugin({
      // suppress react devtools console warning
      __REACT_DEVTOOLS_GLOBAL_HOOK__: '({ isDisabled: true })',
    }),
    new HtmlWebpackPlugin({
      template: './index.html',
      filename: 'index.html',
      inject: 'body',
      favicon: 'assets/logo/favicon.ico',
    }),
  ],
};

module.exports = config;
