<script>
	import { goto } from '$app/navigation';
	import { asset } from '$app/paths';
	import NavButton from '$lib/components/NavButton.svelte';
	import Archive from 'phosphor-svelte/lib/Archive';
	import Folders from 'phosphor-svelte/lib/Folders';
	import Gear from 'phosphor-svelte/lib/Gear';
	import IconContext from 'phosphor-svelte/lib/IconContext';
	import Plus from 'phosphor-svelte/lib/Plus';
	import Queue from 'phosphor-svelte/lib/Queue';
	import { Button, Flexbox, Separator, Text, ThemeRootProvider } from 'svxui';
	import 'svxui/normalize.css';
	import 'svxui/tokens.css';
	import 'svxui/utilities.css';

	let { children } = $props();
</script>

<IconContext values={{ size: '1.4rem' }}>
	<ThemeRootProvider defaultRadius="large">
		<div class="container">
			<aside>
				<Flexbox as="header" gap="1" align="center" justify="start" class="mb-5">
					<img src={asset('/logo.png')} alt="logo" />
					<Flexbox>
						<Text weight="bold" size="5" transform="uppercase" as="div" color="orange">one</Text>
						<Text weight="bold" size="5" transform="uppercase" as="i">fetch</Text>
					</Flexbox>
				</Flexbox>

				<Button
					fullWidth
					size="2"
					variant="solid"
					title="New"
					class="w-100 py-1"
					onclick={() => {
						goto('/');
					}}
				>
					<Plus size="1.2rem" />
					<span>New task</span>
				</Button>

				<Separator size="4" class="my-4" />

				<NavButton route="/active" title="Active downloads">
					<Queue size="1.2rem" />
					<span>Active</span>
				</NavButton>

				<NavButton route="/history" title="Downloads history">
					<Archive size="1.2rem" />
					<span>history</span>
				</NavButton>

				<NavButton route="/files" title="Downloaded files">
					<Folders size="1.2rem" />
					<span>Files</span>
				</NavButton>

				<div class="flex-auto"></div>
				<Separator size="4" class="my-4" />

				<NavButton route="/settings" title="Settings">
					<Gear size="1.2rem" />
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
		--aside-width: 260px;

		width: 100vw;
		min-height: 100vh;
		display: flex;

		aside {
			display: flex;
			flex-direction: column;
			width: var(--aside-width);
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
			width: calc(100vw - var(--aside-width));
			min-height: 100vh;
			padding: var(--space-5);
			background-color: var(--color-background-1);
		}
	}
</style>
