/* eslint-disable */
const webpack = require('webpack');
const HtmlWebpackPlugin = require('html-webpack-plugin');

const config = {
    mode: 'development',
    entry: './index.js',
    output: {
        path: `${__dirname}/public/`,
        filename: 'bundle.js',
    },
    devServer: {
        port: 7900,
        inline: true,
        contentBase: './public',
        publicPath: '/',
    },
    devtool: 'inline-source-map',
    module: {
        rules: [
            {
                test: /\.jsx?$/,
                exclude: '/node_modules/',
                use: [
                    {
                        loader: 'babel-loader',
                        options: {
                            presets: ['es2015', 'react', 'stage-3'],
                        },
                    }
                ],
            },
        ],
    },
    plugins: [
        new webpack.EnvironmentPlugin(['NODE_ENV']),
        new webpack.HotModuleReplacementPlugin(),
        new webpack.DefinePlugin({
            // suppress react devtools console warning
            '__REACT_DEVTOOLS_GLOBAL_HOOK__': '({ isDisabled: true })'
        }),
        new HtmlWebpackPlugin({
            template: './index.html',
            filename: 'index.html',
            inject: 'body',
        })
    ]
};

module.exports = config;
