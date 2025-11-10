<script lang="ts">
	import EmptyState from '$lib/components/EmptyState.svelte';
	import PageLayout from '$lib/components/PageLayout.svelte';
	import ProgressBar from '$lib/components/ProgressBar.svelte';
	import { createDownloadState } from '$lib/state/download-state.svelte';
	import { DownloadStatus } from '$lib/types/types';
	import { formatBytes } from '$lib/utils/format-bytes';
	import { formatProgress } from '$lib/utils/format-progress';
	import { formatRemainingTime } from '$lib/utils/format-remaining-time';
	import { statusColor } from '$lib/utils/status-color';
	import { typeIcons } from '$lib/utils/type-icons';
	import Pause from 'phosphor-svelte/lib/Pause';
	import Play from 'phosphor-svelte/lib/Play';
	import X from 'phosphor-svelte/lib/X';
	import { onMount } from 'svelte';
	import { Badge, Button, Flexbox, Panel, Text } from 'svxui';

	const downloadState = createDownloadState();
	onMount(downloadState.start);
</script>

<PageLayout title="Download" error={downloadState.error}>
	{#if downloadState.downloads?.length > 0}
		<Flexbox direction="column" gap="4">
			{#each downloadState.downloads as dl}
				<!-- Download card -->
				<Panel variant="soft">
					<Flexbox gap="3" direction="column" class="flex-auto">
						<!-- Row 1 => Icon + FileName + Status -->
						<Flexbox gap="3" align="center" class="w-100">
							{@const Icon = typeIcons[dl.type]}
							<Icon size="1.4rem" class="shrink-0" />

							<Text truncate>{dl.customFileName ?? dl.fileName}</Text>
							<div class="flex-auto"></div>

							<Badge variant="soft" size="3" color={statusColor[dl.status]}>
								{dl.status}
							</Badge>
						</Flexbox>

						<!-- Row 2 => progressbar -->
						<Flexbox gap="3" class="w-100">
							<ProgressBar value={dl.progress} color={statusColor[dl.status]} />
						</Flexbox>

						<!-- Row 3 => Stats + Actions -->
						<Flexbox gap="1" justify="between" class="w-100">
							<table class="stats">
								<tbody>
									<tr>
										<td>
											<Text truncate wrap="nowrap">Downloaded</Text>
										</td>
										<td>
											<Flexbox gap="1">
												<Text size="2" truncate wrap="nowrap">
													{formatBytes(Number(dl.downloadedBytes))}
												</Text>
												<Text size="2" muted truncate wrap="nowrap">
													/&nbsp;{formatBytes(Number(dl.fileSize))}
												</Text>
											</Flexbox>
										</td>
									</tr>

									<tr>
										<td>
											<Text truncate wrap="nowrap">Speed</Text>
										</td>
										<td>
											<Flexbox gap="1">
												<Text size="2" truncate wrap="nowrap">
													{formatBytes(dl.speed)}/s
												</Text>
											</Flexbox>
										</td>
									</tr>

									<tr>
										<td>
											<Text truncate wrap="nowrap">Progress</Text>
										</td>
										<td>
											<Text size="2" truncate wrap="nowrap">
												{formatProgress(dl.progress)}
											</Text>
										</td>
									</tr>

									<tr>
										<td>
											<Text truncate wrap="nowrap">Remaining</Text>
										</td>
										<td>
											<Text size="2" truncate wrap="nowrap">
												{formatRemainingTime(
													Number(dl.fileSize),
													Number(dl.downloadedBytes),
													dl.speed ?? 0
												)}
											</Text>
										</td>
									</tr>
								</tbody>
							</table>

							<Flexbox gap="2" align="end">
								{#if dl.status === DownloadStatus.PAUSED}
									<Button
										size="3"
										variant="soft"
										iconOnly
										onclick={() => downloadState.resume(dl.id)}
									>
										<Play weight="bold" />
									</Button>
								{/if}

								{#if dl.status === DownloadStatus.DOWNLOADING}
									<Button
										size="3"
										variant="soft"
										iconOnly
										onclick={() => downloadState.pause(dl.id)}
									>
										<Pause weight="bold" />
									</Button>
								{/if}

								{#if [DownloadStatus.DOWNLOADING, DownloadStatus.PAUSED].includes(dl.status)}
									<Button
										size="3"
										variant="soft"
										iconOnly
										onclick={() => downloadState.cancel(dl.id)}
									>
										<X weight="bold" />
									</Button>
								{/if}
							</Flexbox>
						</Flexbox>
					</Flexbox>
				</Panel>
			{/each}
		</Flexbox>
	{:else}
		<EmptyState text="No downloads..." />
	{/if}
</PageLayout>

<style>
	table {
		&.stats {
			tr {
				td {
					padding: 0;

					&:first-child {
						width: 105px;
					}
				}
			}
		}
	}
</style>
