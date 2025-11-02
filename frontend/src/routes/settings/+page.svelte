<script lang="ts">
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { createSettingsState } from '$lib/state/settings-state.svelte';
	import Check from 'phosphor-svelte/lib/Check';
	import { onMount } from 'svelte';
	import { Button, Flexbox, Input, Panel, Text } from 'svxui';

	const settingsState = createSettingsState();
	onMount(settingsState.get);
</script>

<PageLayout title="Settings" error={settingsState.error}>
	<Flexbox direction="column" gap="6" as="form">
		<Panel variant="soft" size="0">
			<Flexbox direction="column" gap="4" class="p-5">
				<Flexbox gap="4" align="center" as="label">
					<span>1fichier.com API key</span>
					<Input
						class="flex-auto"
						size="3"
						name="apiKey1fichier"
						bind:value={settingsState.apiKey1fichier}
						disabled={settingsState.loading}
					/>
				</Flexbox>

				<Flexbox gap="4" align="center" as="label">
					<span>Jellyfin API key</span>
					<Input
						class="flex-auto"
						size="3"
						name="apiKeyJellyfin"
						bind:value={settingsState.apiKeyJellyfin}
						disabled={settingsState.loading}
					/>
				</Flexbox>
			</Flexbox>
		</Panel>

		<Flexbox>
			<Button
				size="3"
				disabled={settingsState.loading || settingsState.disabled}
				onclick={settingsState.update}
			>
				<Check weight="bold" />
				Save
			</Button>
		</Flexbox>
	</Flexbox>
</PageLayout>

<style>
	span {
		min-width: 145px;
		text-align: right;
	}
</style>
