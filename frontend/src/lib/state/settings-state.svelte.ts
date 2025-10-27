import { api } from '$lib/api/api.svelte';
import { useAsyncState } from '$lib/api/use-async-state.svelte';
import type { Settings } from '$lib/types/types';

export const createSettingsState = () => {
	let apiKey = $state<string>('');
	let settings = $state<Settings | null>(null);

	const asyncStateOptions = {
		immediate: false,
		onSuccess(value: Settings): void {
			settings = value;
			apiKey = value.apiKey ?? apiKey;
		}
	};

	const getState = useAsyncState<Settings, void>(
		api.settings.get, //
		asyncStateOptions
	);
	const updateState = useAsyncState<Settings, Partial<Settings>>(
		api.settings.update, //
		asyncStateOptions
	);

	return {
		get apiKey() {
			return apiKey;
		},
		set apiKey(value) {
			apiKey = value;
		},
		get disabled() {
			return !apiKey || apiKey === settings?.apiKey;
		},
		get loading() {
			return getState.loading || updateState.loading;
		},
		get error() {
			return getState.error || updateState.error;
		},
		get() {
			getState.execute();
		},
		update() {
			updateState.execute({ apiKey });
		}
	};
};
