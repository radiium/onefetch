export interface RemainingTimeResult {
	seconds: number;
	formatted: string;
}

export function formatRemainingTime(
	totalBytes: number,
	downloadedBytes: number,
	speedBytesPerSecond: number
): string {
	if (speedBytesPerSecond <= 0) {
		return '-';
	}

	const remainingBytes = totalBytes - downloadedBytes;
	if (remainingBytes <= 0) {
		return '-';
	}

	const seconds = remainingBytes / speedBytesPerSecond;
	return formatDuration(seconds);
}

function formatDuration(seconds: number): string {
	if (seconds < 60) {
		return `${Math.round(seconds)}s`;
	}

	if (seconds < 3600) {
		const minutes = Math.floor(seconds / 60);
		const secs = Math.round(seconds % 60);
		return `${minutes}min ${secs}s`;
	}

	if (seconds < 86400) {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		return `${hours}h ${minutes}min`;
	}

	const days = Math.floor(seconds / 86400);
	const hours = Math.floor((seconds % 86400) / 3600);
	return `${days}j ${hours}h`;
}
