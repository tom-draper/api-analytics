import svelte from 'rollup-plugin-svelte';
import commonjs from '@rollup/plugin-commonjs';
import typescript from '@rollup/plugin-typescript';
import resolve from '@rollup/plugin-node-resolve';
import livereload from 'rollup-plugin-livereload';
import terser from '@rollup/plugin-terser';
import sveltePreprocess from 'svelte-preprocess';
import json from '@rollup/plugin-json';

const production = !process.env.ROLLUP_WATCH;

export default [
    // Browser bundle
    {
        input: 'src/main.ts',
        output: {
            sourcemap: true,
            format: 'iife',
            name: 'app',
            file: 'public/bundle.js'
        },
        plugins: [
            svelte({
                preprocess: sveltePreprocess({ sourceMap: !production }),
                dev: !production,
                hydratable: true,
                css: (css) => {
                    css.write('bundle.css');
                }
            }),
            json(),
            resolve(),
            commonjs(),
            typescript({
                rootDir: './src',
                sourceMap: !production,
                inlineSources: !production
            }),
            // App.js will be built after bundle.js, so we only need to watch that.
            // By setting a small delay the Node server has a chance to restart before reloading.
            !production &&
                livereload({
                    watch: 'public/App.js',
                    delay: 200
                }),
            production && terser()
        ]
    },
    // Server bundle
    {
        input: 'src/App.svelte',
        output: {
            exports: 'default',
            sourcemap: false,
            format: 'esm',
            name: 'app',
            file: 'public/App.js'
        },
        plugins: [
            svelte({
                preprocess: sveltePreprocess({ sourceMap: !production }),
                generate: 'ssr'
            }),
            json(),
            resolve(),
            commonjs(),
            typescript(),
            production && terser()
        ]
    }
];
