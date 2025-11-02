import { api } from '$lib/api/api.svelte';
import { useAsyncState } from '$lib/api/use-async-state.svelte';
import type { Settings } from '$lib/types/types';

export const createSettingsState = () => {
	let settings = $state<Settings | null>(null);
	let form = $state<Partial<Settings>>({
		apiKey1fichier: '',
		apiKeyJellyfin: ''
	});

	const asyncStateOptions = {
		immediate: false,
		onSuccess(value: Settings): void {
			settings = value;
			form = { ...value };
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
		// API Key 1fichier
		get apiKey1fichier() {
			return form.apiKey1fichier;
		},
		set apiKey1fichier(value) {
			form.apiKey1fichier = value;
		},
		// API Key Jellyfin
		get apiKeyJellyfin() {
			return form.apiKeyJellyfin;
		},
		set apiKeyJellyfin(value) {
			form.apiKeyJellyfin = value;
		},
		// States
		get disabled() {
			return (
				!form.apiKey1fichier ||
				!form.apiKeyJellyfin ||
				(settings?.apiKey1fichier === form.apiKey1fichier &&
					settings?.apiKeyJellyfin === form.apiKeyJellyfin)
			);
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
			updateState.execute($state.snapshot(form));
		}
	};
};
