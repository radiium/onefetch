<script lang="ts">
	import ErrorMessage from '$lib/components/ErrorMessage.svelte';
	import { type Snippet } from 'svelte';
	import { Flex, Text } from 'svxui';

	type Props = {
		title?: string;
		error?: Error | null;
		buttons?: Snippet;
		children?: Snippet;
	};

	let { title, error, buttons, children }: Props = $props();
</script>

<Flex direction="column" gap="3">
	<Flex as="header" align="center" class="mb-5">
		{#if title}
			<Text size="8" weight="bold">{title}</Text>
		{/if}
		<div class="flex-auto"></div>
		{@render buttons?.()}
	</Flex>

	{#if error}
		<ErrorMessage {error} />
	{/if}

	{@render children?.()}
</Flex>
