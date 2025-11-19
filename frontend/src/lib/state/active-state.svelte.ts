import { api } from '$lib/api/api.svelte';
import { useAsyncState, type AsyncStateOptions } from '$lib/api/use-async-state.svelte';
import {
	DownloadStatus,
	type Download,
	type DownloadFilters,
	type DownloadPage,
	type DownloadProgressEvent
} from '$lib/types/types';

export const createActiveState = () => {
	let downloads = $state<Download[]>([]);

	const getAllState = useAsyncState<DownloadPage, DownloadFilters>(api.download.getAll, {
		immediate: false
	});

	const actionOptions: AsyncStateOptions<Download> = {
		immediate: false,
		onSuccess(data) {
			const index = downloads?.findIndex((dl) => dl.id === data.id);
			if (index !== -1) {
				downloads[index] = data;
			}
		}
	};
	const pauseState = useAsyncState<Download, string>(api.download.pause, actionOptions);
	const resumeState = useAsyncState<Download, string>(api.download.resume, actionOptions);
	const cancelState = useAsyncState<Download, string>(api.download.cancel, actionOptions);
	const archiveState = useAsyncState<Download, string>(api.download.archive, actionOptions);

	async function getActiveDownloads() {
		await getAllState.execute({
			status: [
				DownloadStatus.IDLE, //
				DownloadStatus.PENDING,
				DownloadStatus.REQUESTING_INFOS,
				DownloadStatus.REQUESTING_TOKEN,
				DownloadStatus.DOWNLOADING,
				DownloadStatus.PAUSED
			],
			page: 0,
			limit: 1000
		});
		if (getAllState.current) {
			downloads = getAllState.current.data;
		}
	}

	return {
		get downloads() {
			return downloads;
		},
		get loading() {
			return (
				pauseState.loading || //
				resumeState.loading ||
				cancelState.loading ||
				archiveState.loading
			);
		},
		get error() {
			return (
				pauseState.error || //
				resumeState.error ||
				cancelState.error ||
				archiveState.error
			);
		},
		pause: pauseState.execute,
		resume: resumeState.execute,
		cancel: cancelState.execute,
		archive: archiveState.execute,
		start() {
			getActiveDownloads();
			const sse = api.download.streams({
				onConnect() {
					console.log('SSE connected');
				},
				onDisconnect() {
					console.log('SSE disconnected');
				},
				onError(error: Error) {
					console.log('SSE error', error);
				},
				onMessage(event: MessageEvent<string>) {
					const progress = JSON.parse(event.data) as DownloadProgressEvent;
					if (downloads) {
						const index = downloads?.findIndex((dl) => dl.id === progress.downloadId);
						if (index !== -1) {
							downloads[index].status = progress.status;
							downloads[index].fileName = progress.fileName;
							downloads[index].customFileName = progress.customFileName;
							downloads[index].progress = progress.progress;
							downloads[index].speed = progress.speed;
							downloads[index].fileSize = progress.fileSize;
							downloads[index].downloadedBytes = progress.downloadedBytes;
						}
					}
				}
			});

			sse.connect();
			return sse.disconnect;
		}
	};
};
