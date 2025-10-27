<script lang="ts">
	import DownloadStatusBadge from '$lib/components/DownloadStatusBadge.svelte';
	import Dropdown from '$lib/components/Dropdown.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { createHistoryState } from '$lib/state/history-state.svelte';
	import { DownloadStatus, DownloadType } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { formatDate } from '$lib/utils/format-date';
	import { typeIcons } from '$lib/utils/type-icons';
	import ArrowsClockwise from 'phosphor-svelte/lib/ArrowsClockwise';
	import { onMount } from 'svelte';
	import { Button, Flexbox, Panel, Text } from 'svxui';

	const historyState = createHistoryState();
	onMount(historyState.get);
</script>

<PageLayout title="History" error={historyState.error}>
	{#snippet buttons()}
		<Button size="2" disabled={historyState.loading} onclick={historyState.get}>
			<ArrowsClockwise weight="bold" />
			Refresh
		</Button>
	{/snippet}

	<Flexbox direction="column" gap="5">
		<Flexbox gap="3">
			<Dropdown bind:value={historyState.status} options={Object.values(DownloadStatus)}>
				{#snippet buttonContent()}
					STATUS
				{/snippet}
			</Dropdown>

			<Dropdown bind:value={historyState.type} options={Object.values(DownloadType)}>
				{#snippet buttonContent()}
					TYPE
				{/snippet}
			</Dropdown>

			<Button onclick={historyState.resetFilters} disabled={!historyState.hasFilters}>
				Reset filters
			</Button>
		</Flexbox>

		{#if historyState.current}
			<Flexbox direction="column" gap="2">
				{#each historyState.current?.data as dl}
					<Panel variant="soft">
						<Flexbox direction="column" gap="3">
							<Flexbox gap="3">
								{@const Icon = typeIcons[dl.type]}
								<Icon size="1.4rem" class="shrink-0" />

								<Flexbox gap="2" direction="column" class="flex-auto min-w-0">
									<Text truncate>{dl.fileName}</Text>

									<Flexbox gap="1" class="min-w-0">
										<Text size="2" truncate wrap="nowrap">{formatBytes(Number(dl.fileSize))}</Text>
										<Text size="2" truncate wrap="nowrap" muted>-</Text>
										<Text size="2" truncate wrap="nowrap" muted>{formatDate(dl.createdAt)}</Text>
									</Flexbox>
								</Flexbox>

								<Flexbox gap="2" direction="column" align="end">
									<DownloadStatusBadge {dl} />
								</Flexbox>
							</Flexbox>
						</Flexbox>
					</Panel>
				{/each}
			</Flexbox>

			{#if historyState.current.pagination.totalPages > 1}
				<Flexbox gap="3">
					{#each { length: historyState.current.pagination.totalPages } as _, i}
						{@const page = i + 1}
						<Button
							size="2"
							iconOnly
							variant={historyState.page === page ? 'solid' : 'outline'}
							onclick={() => {
								console.log('page', page);
								historyState.page = page as number;
							}}
						>
							{page}
						</Button>
					{/each}
				</Flexbox>
			{/if}
		{:else}
			<EmptyState text="No downloads..." />
		{/if}
	</Flexbox>
</PageLayout>
