import { includeIgnoreFile } from '@eslint/compat';
import js from '@eslint/js';
import prettier from 'eslint-config-prettier';
import svelte from 'eslint-plugin-svelte';
import { defineConfig } from 'eslint/config';
import globals from 'globals';
import { fileURLToPath } from 'node:url';
import ts from 'typescript-eslint';
import svelteConfig from './svelte.config.js';

const gitignorePath = fileURLToPath(new URL('./.gitignore', import.meta.url));

export default defineConfig(
	includeIgnoreFile(gitignorePath),
	js.configs.recommended,
	...ts.configs.recommended,
	...svelte.configs.recommended,
	prettier,
	...svelte.configs.prettier,
	{
		languageOptions: {
			globals: { ...globals.browser, ...globals.node }
		}
	},
	{
		files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
		languageOptions: {
			parserOptions: {
				projectService: true,
				extraFileExtensions: ['.svelte'],
				parser: ts.parser,
				svelteConfig
			}
		},
        rules: {
            'svelte/prefer-const': 'error',
            'svelte/no-unused-svelte-ignore': 'warn',
            'svelte/no-unused-props': 'warn',
            'svelte/no-navigation-without-resolve': [
                'warn',
                {
                    ignoreGoto: false,
                    ignoreLinks: true,
                    ignorePushState: false,
                    ignoreReplaceState: false
                }
            ]
        }
	},
	{
        rules: {
            'no-undef': 'off',
            '@typescript-eslint/no-inferrable-types': 'off',
            '@typescript-eslint/no-unused-vars': 'warn',
            '@typescript-eslint/no-unused-expressions': 'warn',
            '@typescript-eslint/no-explicit-any': 'warn',
            '@typescript-eslint/no-non-null-assertion': 'warn',
            'no-case-declarations': 'off',
            'no-console': ['warn', { allow: ['warn', 'error'] }],
            'prefer-const': 'off',
            'no-var': 'error'
        }
    },
);
