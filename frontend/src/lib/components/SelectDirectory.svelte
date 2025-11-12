<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Button, Flexbox, Floating } from 'svxui';

	type Props = {
		options: string[];
		disabled?: boolean;
		onSelect?: (value: string) => void;
		children: Snippet;
	};
	let { options = [], disabled, onSelect, children }: Props = $props();

	let isOpen = $state(false);
</script>

<Floating
	size="2"
	variant="outline"
	placement="bottom-end"
	offset={4}
	autoUpdate
	flip
	closeOnClickOutside
	bind:isOpen
>
	{#snippet trigger()}
		<Button size="3" iconOnly variant="outline" onclick={() => (isOpen = !isOpen)} {disabled}>
			{@render children?.()}
		</Button>
	{/snippet}
	{#snippet content()}
		<Flexbox direction="column" gap="1" align="center" style="min-width: 250px">
			{#each options as opt}
				<Button
					variant="clear"
					align="start"
					fullWidth
					onclick={() => onSelect?.(opt)}
					style="--button-background-hover: var(--accent-5);"
				>
					{opt}
				</Button>
			{/each}
		</Flexbox>
	{/snippet}
</Floating>
