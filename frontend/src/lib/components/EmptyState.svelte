<script lang="ts">
	import { Flex, Text } from 'svxui';

	type Props = {
		text?: string;
	};

	let { text }: Props = $props();
</script>

<Flex justify="center" direction="column" align="center" class="py-9 w-100" gap="5">
	{#if text}
		<Text size="7" weight="medium" align="center" muted as="i">{text}</Text>
	{/if}
	<img src="/doc.svg" alt="" />
</Flex>

<style>
	img {
		max-width: min(50%, 200px);
	}
</style>
