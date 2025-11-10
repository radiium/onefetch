<script>
	import { goto } from '$app/navigation';
	import { asset } from '$app/paths';
	import NavButton from '$lib/components/NavButton.svelte';
	import Archive from 'phosphor-svelte/lib/Archive';
	import BoxArrowDown from 'phosphor-svelte/lib/BoxArrowDown';
	import CheckCircle from 'phosphor-svelte/lib/CheckCircle';
	import Folders from 'phosphor-svelte/lib/Folders';
	import Gear from 'phosphor-svelte/lib/Gear';
	import IconContext from 'phosphor-svelte/lib/IconContext';
	import Pause from 'phosphor-svelte/lib/Pause';
	import Play from 'phosphor-svelte/lib/Play';
	import Plus from 'phosphor-svelte/lib/Plus';
	import Stop from 'phosphor-svelte/lib/Stop';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import { Button, Flexbox, Separator, Text, ThemeRootProvider } from 'svxui';
	import 'svxui/normalize.css';
	import 'svxui/tokens.css';
	import 'svxui/utilities.css';

	let { children } = $props();
</script>

<IconContext values={{ size: '1.4rem' }}>
	<ThemeRootProvider>
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

				<NavButton route="/dl/all" title="All">
					<BoxArrowDown size="1.2rem" />
					<span>All</span>
				</NavButton>
				<NavButton route="/dl/downloading" title="Downloading">
					<Play size="1.2rem" />
					<span>Downloading</span>
				</NavButton>
				<NavButton route="/dl/paused" title="Paused">
					<Pause size="1.2rem" />
					<span>Paused</span>
				</NavButton>
				<NavButton route="/dl/cancelled" title="Cancelled">
					<Stop size="1.2rem" />
					<span>Cancelled</span>
				</NavButton>
				<NavButton route="/dl/failed" title="Failed">
					<WarningCircle size="1.2rem" />
					<span>Failed</span>
				</NavButton>
				<NavButton route="/dl/completed" title="Completed">
					<CheckCircle size="1.2rem" />
					<span>Completed</span>
				</NavButton>

				<Separator size="4" class="my-4" />
				<NavButton route="/files" title="Files">
					<Folders size="1.2rem" />
					<span>Files</span>
				</NavButton>

				<Separator size="4" class="my-4" />

				<NavButton route="/download" title="Download">
					<BoxArrowDown size="1.2rem" />
					<span>Download</span>
				</NavButton>

				<NavButton route="/history" title="History">
					<Archive size="1.2rem" />
					<span>History</span>
				</NavButton>

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
