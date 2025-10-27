<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { Snippet } from 'svelte';
	import { Button, Flexbox } from 'svxui';

	type Props = {
		route?: string;
		title?: string;
		onclick?: () => void;
		children?: Snippet;
	};

	let { route, title, onclick, children }: Props = $props();

	let active = $derived(route && page.url.pathname === `${route}`);
</script>

<Button
	fullWidth
	size="3"
	color={active ? 'neutral' : 'neutral'}
	variant={active ? 'soft' : 'clear'}
	class={active ? 'nav-button active' : 'nav-button'}
	align="start"
	{title}
	onclick={() => {
		if (route) {
			goto(route);
		}
		onclick?.();
	}}
>
	<Flexbox align="center" gap="3" class="w-100 py-1">
		{@render children?.()}
	</Flexbox>
</Button>

<style>
	:global(.button.nav-button.active:after) {
		content: '➜';
		position: absolute;
		right: var(--space-3);
	}
</style>
