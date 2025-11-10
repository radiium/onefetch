<script lang="ts">
	import { HttpError } from '$lib/api/http-client';
	import WarningCircle from 'phosphor-svelte/lib/WarningCircle';
	import { Flexbox, Panel, Text } from 'svxui';

	type Props = {
		error?: Error;
	};

	let { error }: Props = $props();
	let name = $derived(error?.name ?? 'Error');
	let cause = $derived(error?.cause ?? '');
	let message = $derived(error?.message ?? '');
	let extra = $derived(error instanceof HttpError ? (error?.data as any)?.error : '');
</script>

<Panel variant="soft" color="red" size="6" style="padding: var(--space-4) ">
	<Flexbox gap="3">
		<WarningCircle size="24px" weight="fill" color="var(--tomato-9)" class="shrink-0" />

		<Flexbox direction="column" gap="1">
			<Text color="red" weight="bold">{name}</Text>

			{#if cause}
				<Text color="red" size="2">{cause}</Text>
			{/if}

			{#if message}
				<Text color="red" size="2">{message}</Text>
			{/if}

			{#if extra}
				<Text color="red" size="2">{extra}</Text>
			{/if}
		</Flexbox>
	</Flexbox>
</Panel>
