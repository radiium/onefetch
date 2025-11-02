import { DownloadStatus } from '$lib/types/types';
import type { Color } from 'svxui';

export const statusColor: Record<DownloadStatus, Color> = {
	[DownloadStatus.PAUSED]: 'orange',
	[DownloadStatus.CANCELLED]: 'yellow',
	[DownloadStatus.DOWNLOADING]: 'green',
	[DownloadStatus.COMPLETED]: 'blue',
	[DownloadStatus.FAILED]: 'red',
	[DownloadStatus.REQUESTING]: 'neutral',
	[DownloadStatus.PENDING]: 'neutral'
};
