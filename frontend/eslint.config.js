import eslintPluginSvelte from 'eslint-plugin-svelte';

export default [
	...eslintPluginSvelte.configs['flat/recommended'],
	{
		ignores: [
			'.svelte-kit/**/*',
			'build/**/*',
			'node_modules/**/*',
			'refactor.js'
		]
	},
	{
		rules: {
            'no-unused-vars': 'warn',
		    'no-console': ['warn', { allow: ['warn', 'error', 'info'] }],
		    'eqeqeq': ['warn', 'always'],
		    'curly': 'warn',
		    'svelte/require-each-key': 'warn',
		    'svelte/no-navigation-without-resolve': 'warn',
            'svelte/prefer-svelte-reactivity': 'warn'
		}
	}
];
