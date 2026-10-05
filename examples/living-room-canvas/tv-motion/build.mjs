import { build } from 'esbuild';
import { createRequire } from 'node:module';
import { dirname, resolve } from 'node:path';
const require = createRequire(import.meta.url);
const { create } = require('enhanced-resolve');
const browserResolve = create.sync({extensions: ['.js', '.mjs', '.json'],
  conditionNames: ['browser', 'import', 'default'], mainFields: ['browser', 'module', 'main']});

// Resolve this project's npm dependencies with browser/ESM conditions rather
// than inheriting an unrelated Yarn Plug'n'Play manifest from the Mac's home.
const npmResolution = {
  name: 'local-browser-resolution',
  setup(builder) {
    builder.onResolve({filter: /^[^./]/}, args => ({
      path: browserResolve(dirname(args.importer || resolve('package.json')), args.path),
    }));
  },
};
await build({entryPoints: ['src/player.tsx'], bundle: true,
  outfile: '../web/motion/player.js', minify: true, sourcemap: false,
  plugins: [npmResolution], define: {'process.env.NODE_ENV': '"production"'},
  jsx: 'automatic', target: 'es2022'});
