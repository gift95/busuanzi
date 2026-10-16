const path = require('path');

module.exports = {
    entry: {
        'busuanzi': './busuanzi.ts',
        'busuanzi.jsonp': './busuanzi.jsonp.ts',
        'busuanzi.lite': './busuanzi.lite.ts',
        'busuanzi.pjax.lite': './busuanzi.pjax.lite.ts',
    },
    output: {
        filename: '[name].js',
        path: path.resolve(__dirname, './')
    },
    mode: 'production',
    module: {
        rules: [{
            test: /\.tsx?$/,
            use: 'ts-loader',
            exclude: /node_modules/
        }]
    },
    resolve: {
        extensions: ['.ts']
    },
}
