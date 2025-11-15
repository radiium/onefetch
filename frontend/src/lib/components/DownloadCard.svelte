<script lang="ts">
	import type { Download } from '$lib/types/types';
	import { DownloadStatus } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { formatProgress } from '$lib/utils/format-progress';
	import { formatRemainingTime } from '$lib/utils/format-remaining-time';
	import { statusColor } from '$lib/utils/status-color';
	import { typeIcons } from '$lib/utils/type-icons';
	import Pause from 'phosphor-svelte/lib/Pause';
	import Play from 'phosphor-svelte/lib/Play';
	import X from 'phosphor-svelte/lib/X';
	import { Badge, Button, Flexbox, Panel, Text } from 'svxui';
	import ProgressBar from './ProgressBar.svelte';

	type Props = {
		download: Download;
		pause?: (id: string) => void;
		resume?: (id: string) => void;
		cancel?: (id: string) => void;
	};

	let { download, pause, resume, cancel }: Props = $props();

	let fileName = $derived(download.customFileName ?? download.fileName);
	let color = $derived(statusColor[download.status]);
	let Icon = $derived(typeIcons[download?.type]);
</script>

{#if download}
	<Panel variant="soft">
		<Flexbox gap="3" direction="column" class="flex-auto">
			<!-- Row 1 => Icon + FileName + Status -->
			<Flexbox gap="3" align="center" class="w-100">
				<Icon size="1.4rem" class="shrink-0" />

				<Text truncate>{fileName}</Text>
				<div class="flex-auto"></div>

				<Badge variant="soft" size="3" {color}>
					{download.status}
				</Badge>
			</Flexbox>

			<!-- Row 2 => progressbar -->
			<Flexbox gap="3" class="w-100">
				<ProgressBar value={download.progress} {color} />
			</Flexbox>

			<!-- Row 3 => Stats + Actions -->
			<Flexbox gap="1" justify="between" class="w-100">
				<table class="stats">
					<tbody>
						<tr>
							<td>
								<Text truncate wrap="nowrap">Progress</Text>
							</td>
							<td>
								<Text size="2" truncate wrap="nowrap">
									{formatProgress(download.progress)}
								</Text>
							</td>
						</tr>

						<tr>
							<td>
								<Text truncate wrap="nowrap">Speed</Text>
							</td>
							<td>
								<Flexbox gap="1">
									<Text size="2" truncate wrap="nowrap">
										{formatBytes(download.speed)}/s
									</Text>
								</Flexbox>
							</td>
						</tr>

						<tr>
							<td>
								<Text truncate wrap="nowrap">Remaining</Text>
							</td>
							<td>
								<Text size="2" truncate wrap="nowrap">
									{formatRemainingTime(
										Number(download.fileSize),
										Number(download.downloadedBytes),
										download.speed ?? 0
									)}
								</Text>
							</td>
						</tr>

						<tr>
							<td>
								<Text truncate wrap="nowrap">Downloaded</Text>
							</td>
							<td>
								<Flexbox gap="1">
									<Text size="2" truncate wrap="nowrap">
										{formatBytes(Number(download.downloadedBytes))}
									</Text>
									<Text size="2" muted truncate wrap="nowrap">
										/&nbsp;{formatBytes(Number(download.fileSize))}
									</Text>
								</Flexbox>
							</td>
						</tr>
					</tbody>
				</table>

				<Flexbox gap="2" align="end">
					{#if download.status === DownloadStatus.PAUSED}
						<Button size="3" variant="soft" iconOnly onclick={() => resume?.(download.id)}>
							<Play weight="bold" />
						</Button>
					{/if}

					{#if download.status === DownloadStatus.DOWNLOADING}
						<Button size="3" variant="soft" iconOnly onclick={() => pause?.(download.id)}>
							<Pause weight="bold" />
						</Button>
					{/if}

					{#if [DownloadStatus.DOWNLOADING, DownloadStatus.PAUSED].includes(download.status)}
						<Button size="3" variant="soft" iconOnly onclick={() => cancel?.(download.id)}>
							<X weight="bold" />
						</Button>
					{/if}
				</Flexbox>
			</Flexbox>
		</Flexbox>
	</Panel>
{/if}
