import { api } from '$lib/api/api.svelte';
import { useAsyncState } from '$lib/api/use-async-state.svelte';
import type { FSNode } from '$lib/types/types';

export const createFilesState = () => {
	let data = $state<FSNode | null>(null);

	const asyncStateOptions = {
		immediate: false,
		onSuccess(value: FSNode): void {
			data = value;
		}
	};

	const getState = useAsyncState<FSNode, void>(
		api.files.get, //
		asyncStateOptions
	);

	const createState = useAsyncState<FSNode, { path?: string; dirname?: string }>(
		api.files.create, //
		asyncStateOptions
	);

	const deleteState = useAsyncState<FSNode, string>(
		api.files.delete, //
		asyncStateOptions
	);

	return {
		get data() {
			return data;
		},
		// States
		get loading() {
			return getState.loading || createState.loading || deleteState.loading;
		},
		get error() {
			return getState.error || createState.error || deleteState.error;
		},
		get() {
			getState.execute();
		},
		create(path?: string, dirname?: string) {
			createState.execute({ path, dirname });
		},
		delete(path?: string) {
			deleteState.execute(path);
		}
	};
};
