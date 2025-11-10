<script lang="ts">
	import DownloadStatusBadge from '$lib/components/DownloadStatusBadge.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { DownloadStatus, DownloadType } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { formatDate } from '$lib/utils/format-date';
	import { typeIcons } from '$lib/utils/type-icons';
	import { onMount } from 'svelte';
	import { Button, Flexbox, Panel, Text } from 'svxui';
	import { createDownloadListState } from './download-list-state.svelte';

	type Props = {
		status: DownloadStatus[];
		type: DownloadType[];
	};

	let { status, type }: Props = $props();

	const list = createDownloadListState({ status, type });
	onMount(list.get);
</script>

<Flexbox direction="column" gap="5">
	{#if Array.isArray(list.current?.data) && list.current?.data?.length}
		<Flexbox direction="column" gap="2">
			{#each list.current?.data as dl}
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

		{#if list.current.pagination.totalPages > 1}
			<Flexbox gap="3">
				{#each { length: list.current.pagination.totalPages } as _, i}
					{@const page = i + 1}
					<Button
						size="2"
						iconOnly
						variant={list.page === page ? 'solid' : 'outline'}
						onclick={() => {
							console.log('page', page);
							list.page = page as number;
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
