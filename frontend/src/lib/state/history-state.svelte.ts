import { api } from '$lib/api/api.svelte';
import { useAsyncState } from '$lib/api/use-async-state.svelte';
import {
	DownloadStatus,
	type DownloadFilters,
	type DownloadPage,
	type DownloadType
} from '$lib/types/types';

export const createHistoryState = () => {
	let status = $state<DownloadStatus[]>([]);
	let type = $state<DownloadType[]>([]);
	let page = $state<number>(1);
	let limit = $state<number>(10);

	const historyState = useAsyncState<DownloadPage, DownloadFilters>(api.download.getAll, {
		immediate: false
	});

	function get() {
		historyState.execute({
			status,
			type,
			page,
			limit
		});
	}

	return {
		// Status
		get status() {
			return status;
		},
		set status(value) {
			console.log('status', value);
			status = value;
			page = 0;
			get();
		},
		// Type
		get type() {
			return type;
		},
		set type(value) {
			console.log('type', value);
			type = value;
			page = 0;
			get();
		},

		get hasFilters() {
			return status?.length > 0 || type?.length > 0;
		},

		// Page
		get page() {
			return page;
		},
		set page(value) {
			page = value;
			get();
		},
		// Limit
		get limit() {
			return limit;
		},
		set limit(value) {
			limit = value;
			get();
		},
		// State
		get current() {
			return historyState.current;
		},
		get loading() {
			return historyState.loading;
		},
		get error() {
			return historyState.error;
		},
		get,
		resetFilters() {
			status = [];
			type = [];
			page = 0
			get()
		}
	};
};
