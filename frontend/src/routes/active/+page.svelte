<script lang="ts">
	import DownloadCard from '$lib/components/DownloadCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { createActiveState } from '$lib/state/active-state.svelte';
	import { onMount } from 'svelte';
	import { Flexbox } from 'svxui';

	const activeState = createActiveState();
	onMount(activeState.start);
</script>

<PageLayout title="Download" error={activeState.error}>
	{#if activeState.downloads?.length > 0}
		<Flexbox direction="column" gap="4">
			{#each activeState.downloads as download}
				<DownloadCard
					{download}
					pause={activeState.pause}
					resume={activeState.resume}
					cancel={activeState.cancel}
				/>
			{/each}
		</Flexbox>
	{:else}
		<EmptyState text="No downloads in progress..." />
	{/if}
</PageLayout>
