import { DownloadStatus } from '$lib/types/types';
import type { Color } from 'svxui';

export function statusColor(status: DownloadStatus): Color {
	switch (status) {
		case DownloadStatus.PAUSED:
			return 'orange';
		case DownloadStatus.CANCELLED:
			return 'yellow';
		case DownloadStatus.DOWNLOADING:
			return 'green';
		case DownloadStatus.COMPLETED:
			return 'blue';
		case DownloadStatus.FAILED:
			return 'red';
		case DownloadStatus.REQUESTING:
		case DownloadStatus.PENDING:
		default:
			return 'neutral';
	}
}
