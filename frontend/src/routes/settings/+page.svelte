<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { createSettingsState } from '$lib/state/settings-state.svelte';
	import Check from 'phosphor-svelte/lib/Check';
	import { onMount } from 'svelte';
	import { Button, Flexbox, Input, Text } from 'svxui';

	const settingsState = createSettingsState();
	onMount(settingsState.get);
</script>

<PageLayout title="Settings" error={settingsState.error}>
	{#snippet buttons()}
		<Button
			size="2"
			disabled={settingsState.loading || settingsState.disabled}
			onclick={settingsState.update}
		>
			<Check weight="bold" />
			Save
		</Button>
	{/snippet}

	<Flexbox direction="column" gap="6" as="form">
		<Flexbox as="label" direction="column" gap="2">
			<Flexbox align="center" gap="2">
				<Text>API key</Text>
				<Text muted as="i">1fichier.com</Text>
			</Flexbox>
			<Input
				size="3"
				name="apiKey"
				placeholder="Your API key"
				bind:value={settingsState.apiKey}
				disabled={settingsState.loading}
			/>
		</Flexbox>
	</Flexbox>
</PageLayout>
