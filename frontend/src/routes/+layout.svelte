<script>
	import NavButton from '$lib/components/NavButton.svelte';
	import Archive from 'phosphor-svelte/lib/Archive';
	import BoxArrowDown from 'phosphor-svelte/lib/BoxArrowDown';
	import Gear from 'phosphor-svelte/lib/Gear';
	import IconContext from 'phosphor-svelte/lib/IconContext';
	import { Flexbox, Text, ThemeRootProvider } from 'svxui';
	import 'svxui/normalize.css';
	import 'svxui/tokens.css';
	import 'svxui/utilities.css';

	let { children } = $props();
</script>

<IconContext values={{ size: '1rem' }}>
	<ThemeRootProvider>
		<div class="container">
			<aside>
				<Flexbox as="header" gap="1" align="center" justify="start" class="mb-5">
					<img src="logo.png" alt="logo" />
					<Flexbox>
						<Text weight="bold" size="5" transform="uppercase" as="div" color="orange">one</Text>
						<Text weight="bold" size="5" transform="uppercase" as="i">fetch</Text>
					</Flexbox>
				</Flexbox>

				<NavButton route="/" title="Download">
					<BoxArrowDown />
					<span>Download</span>
				</NavButton>

				<NavButton route="/history" title="History">
					<Archive />
					<span>History</span>
				</NavButton>

				<NavButton route="/settings" title="Settings">
					<Gear />
					<span>Settings</span>
				</NavButton>
			</aside>
			<main>
				{@render children?.()}
			</main>
		</div>
	</ThemeRootProvider>
</IconContext>

<style>
	.container {
		width: 100vw;
		min-height: 100vh;
		display: flex;

		aside {
			display: flex;
			flex-direction: column;
			width: 180px;
			height: 100vh;
			gap: var(--space-1);
			padding: var(--space-3);
			background-color: var(--color-background-0);
			flex-shrink: 0;

			img {
				margin-top: 3px;
				width: 32px;
				height: auto;
			}
		}

		main {
			flex: 1 1 auto;
			width: calc(100vw - 180px);
			min-height: 100vh;
			padding: var(--space-5);
			background-color: var(--color-background-1);
		}
	}
</style>
