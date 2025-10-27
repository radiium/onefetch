import { api } from '$lib/api/api.svelte';
import { useAsyncState, type AsyncStateOptions } from '$lib/api/use-async-state.svelte';
import {
	DownloadStatus,
	DownloadType,
	type CreateDownloadInput,
	type Download,
	type DownloadFilters,
	type DownloadPage,
	type DownloadProgressEvent
} from '$lib/types/types';

export const createDownloadState = () => {
	let downloads = $state<Download[]>([]);
	let url = $state<string>('https://1fichier.com/?tqz6yz5fcm1pz1tnuzh8');
	let type = $state<DownloadType>(DownloadType.MOVIE);

	const createState = useAsyncState<Download, CreateDownloadInput>(
		api.download.create, //
		{
			immediate: false,
			onSuccess(data) {
				downloads = [...downloads, data];
			}
		}
	);

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

	const historyState = useAsyncState<DownloadPage, DownloadFilters>(api.download.getAll, {
		immediate: false
	});

	async function getActiveDownloads() {
		await historyState.execute({
			status: [
				DownloadStatus.PENDING, //
				DownloadStatus.REQUESTING,
				DownloadStatus.DOWNLOADING,
				DownloadStatus.PAUSED
			],
			page: 0,
			limit: 1000
		});
		if (historyState.current) {
			downloads = historyState.current.data;
		}
	}

	return {
		// Url
		get url() {
			return url;
		},
		set url(value) {
			url = value;
		},
		// Type
		get type() {
			return type;
		},
		set type(value) {
			type = value;
		},
		//
		get downloads() {
			return downloads;
		},
		//
		get loading() {
			return (
				createState.loading ||
				pauseState.loading ||
				resumeState.loading ||
				cancelState.loading ||
				archiveState.loading
			);
		},
		get error() {
			return (
				createState.error ||
				pauseState.error ||
				resumeState.error ||
				cancelState.error ||
				archiveState.error
			);
		},
		create() {
			createState.execute({ url, type });
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
					console.log('SSE progress', event);
					const progress = JSON.parse(event.data) as DownloadProgressEvent;

					if (downloads) {
						const index = downloads?.findIndex((dl) => dl.id === progress.downloadId);
						if (index !== -1) {
							downloads[index].status = progress.status;
							downloads[index].fileName = progress.fileName;
							downloads[index].progress = progress.progress;
							downloads[index].speed = progress.speed;
							downloads[index].fileSize = progress.fileSize;
							downloads[index].downloadedBytes = progress.downloadedBytes;
						}
					}
				}
			});

			sse.connect();
			return () => sse.disconnect();
		}
	};
};
