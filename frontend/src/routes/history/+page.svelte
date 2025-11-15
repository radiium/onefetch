<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import DownloadPagination from '$lib/components/DownloadPagination.svelte';
	import DownloadStatusBadge from '$lib/components/DownloadStatusBadge.svelte';
	import DropdownFilter from '$lib/components/DropdownFilter.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import PageLayout from '$lib/components/PageLayout.svelte';
	import { createHistoryState } from '$lib/state/history-state.svelte';
	import { DownloadStatus, DownloadType } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { formatDate } from '$lib/utils/format-date';
	import { typeIcons } from '$lib/utils/type-icons';
	import ArrowsClockwise from 'phosphor-svelte/lib/ArrowsClockwise';
	import { onMount } from 'svelte';
	import { Button, Flexbox, isBrowser, Panel, Text } from 'svxui';

	const historyState = createHistoryState();
	onMount(historyState.get);

	function parseParam<T>(param?: string | null): T[] {
		return (param?.split(',').filter(Boolean) ?? []) as T[];
	}

	onMount(() => {
		historyState.status = parseParam<DownloadStatus>(page.url.searchParams.get('status'));
		historyState.type = parseParam<DownloadType>(page.url.searchParams.get('type'));
	});

	$effect(() => {
		if (isBrowser()) {
			const searchParams = new URLSearchParams();
			searchParams.set('status', historyState.status?.join(','));
			searchParams.set('type', historyState.type?.join(','));
			goto(`?${searchParams.toString()}`, {
				replaceState: true,
				noScroll: true,
				keepFocus: true
			});
		}
	});
</script>

<PageLayout title="History" error={historyState.error}>
	<Flexbox direction="column" gap="5">
		<Flexbox gap="3">
			<DropdownFilter
				name="Status"
				options={Object.values(DownloadStatus)}
				bind:value={historyState.status}
			/>

			<DropdownFilter
				name="Type"
				options={Object.values(DownloadType)}
				bind:value={historyState.type}
			/>

			<Button
				variant="outline"
				onclick={historyState.resetFilters}
				disabled={!historyState.hasFilters}
			>
				Reset filters
			</Button>

			<div class="flex-auto"></div>

			<Button disabled={historyState.loading} onclick={historyState.get}>
				<ArrowsClockwise weight="bold" />
				Refresh
			</Button>
		</Flexbox>

		{#if Array.isArray(historyState.current?.data) && historyState.current?.data?.length}
			<Flexbox direction="column" gap="2">
				{#each historyState.current?.data as dl}
					<Panel variant="soft">
						<Flexbox direction="column" gap="3">
							<Flexbox gap="3">
								{@const Icon = typeIcons[dl.type]}
								<Icon size="1.4rem" class="shrink-0" />

								<Flexbox gap="2" direction="column" class="flex-auto min-w-0">
									<Text truncate>{dl.customFileName ?? dl.fileName}</Text>

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
				<DownloadPagination
					bind:currentPage={historyState.page}
					pagination={historyState.current.pagination}
				/>
			{/if}
		{:else}
			<EmptyState text="No downloads..." />
		{/if}
	</Flexbox>
</PageLayout>
