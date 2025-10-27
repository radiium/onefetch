export function formatProgress(progress?: number): string {
	if (!progress) return '';

	if (progress === 100) {
		return '100%';
	}

	return progress.toFixed(2) + '%';
}
