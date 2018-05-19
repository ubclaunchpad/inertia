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
        test: /\.sass/,
        exclude: /node_modules/,
        use: ['style-loader', 'css-loader', 'sass-loader'],
      }
    ],
  },
  plugins: [
    new webpack.DefinePlugin({
      'process.env': {
        // define environment variables here
        NODE_ENV: JSON.stringify(process.env.NODE_ENV),
      },
    }),
    new webpack.DefinePlugin({
      // suppress react devtools console warning
      __REACT_DEVTOOLS_GLOBAL_HOOK__: '({ isDisabled: true })',
    }),
    new HtmlWebpackPlugin({
      template: './index.html',
      filename: 'index.html',
      inject: 'body',
    }),
  ],
};

module.exports = config;
