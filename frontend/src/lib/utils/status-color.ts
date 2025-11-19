import { DownloadStatus } from '$lib/types/types';
import type { Color } from 'svxui';

export const statusColor: Record<DownloadStatus, Color> = {
	[DownloadStatus.PAUSED]: 'orange',
	[DownloadStatus.CANCELLED]: 'yellow',
	[DownloadStatus.DOWNLOADING]: 'green',
	[DownloadStatus.COMPLETED]: 'blue',
	[DownloadStatus.FAILED]: 'red',
	[DownloadStatus.REQUESTING_INFOS]: 'neutral',
	[DownloadStatus.REQUESTING_TOKEN]: 'neutral',
	[DownloadStatus.IDLE]: 'neutral',
	[DownloadStatus.PENDING]: 'neutral'
};
