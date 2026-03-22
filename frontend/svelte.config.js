import adapter from '@sveltejs/adapter-static';
import { sveltePhosphorOptimize } from 'phosphor-svelte/vite';
import { sveltePreprocess } from 'svelte-preprocess';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: [sveltePreprocess(), sveltePhosphorOptimize()],
	kit: { adapter: adapter() }
};

export default config;
