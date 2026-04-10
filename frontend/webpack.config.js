const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');

module.exports = {
  entry: './src/index.tsx',
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'bundle.js',
    publicPath: '/'
  },
  resolve: {
    extensions: ['.ts', '.tsx', '.js', '.jsx']
  },
  module: {
    rules: [
      {
        test: /\.(ts|tsx)$/,
        exclude: /node_modules/,
        use: 'ts-loader'
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader']
      }
    ]
  },
  plugins: [
    new HtmlWebpackPlugin({
      template: './public/index.html'
    })
  ],
  devServer: {
    port: 3001,
    historyApiFallback: true,
    proxy: [
      {
        context: ['/api'],
        target: 'http://localhost:8080',
        changeOrigin: true,
        proxyTimeout: 0,   // no timeout — required for SSE long-lived connections
        timeout: 0,
        secure: false,
        // http-proxy-middleware v2 (used by webpack-dev-server v4) callback syntax:
        onProxyRes(proxyRes) {
          // Remove Content-Length so the proxy streams instead of buffers.
          // Without this, http-proxy waits for a "complete" response that never comes.
          delete proxyRes.headers['content-length'];
        },
      }
    ]
  }
};
