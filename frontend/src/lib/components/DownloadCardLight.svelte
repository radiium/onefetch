<script lang="ts">
	import type { Download } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { formatDate } from '$lib/utils/format-date';
	import { typeIcons } from '$lib/utils/type-icons';
	import { Flexbox, Panel, Text } from 'svxui';
	import DownloadStatusBadge from './DownloadStatusBadge.svelte';

	type Props = {
		download: Download;
	};

	let { download }: Props = $props();

	let Icon = $derived(typeIcons[download?.type]);
</script>

{#if download}
	<Panel variant="soft">
		<Flexbox direction="column" gap="3">
			<Flexbox gap="3">
				<Icon size="1.4rem" class="shrink-0" />

				<Flexbox gap="2" direction="column" class="flex-auto min-w-0">
					<Text truncate>{download.customFileName ?? download.fileName}</Text>

					<Flexbox gap="1" class="min-w-0">
						<Text size="2" truncate wrap="nowrap">{formatBytes(Number(download.fileSize))}</Text>
						<Text size="2" truncate wrap="nowrap" muted>-</Text>
						<Text size="2" truncate wrap="nowrap" muted>{formatDate(download.createdAt)}</Text>
					</Flexbox>
				</Flexbox>

				<Flexbox gap="2" direction="column" align="end">
					<DownloadStatusBadge dl={download} />
				</Flexbox>
			</Flexbox>
		</Flexbox>
	</Panel>
{/if}
