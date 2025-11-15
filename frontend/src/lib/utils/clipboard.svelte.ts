import { browser } from '$app/environment';

export function useClipboard() {
	let text = $state<string>('');
	let error = $state<Error | null>(null);
	const isSupported = $state(
		browser && typeof navigator !== 'undefined' && 'clipboard' in navigator
	);

	async function copy(value: string): Promise<boolean> {
		if (!isSupported) {
			error = new Error('Clipboard API not supported');
			console.warn(error);
			return false;
		}

		try {
			await navigator.clipboard.writeText(value);
			text = value;
			error = null;
			return true;
		} catch (e) {
			console.error(e);
			error = e instanceof Error ? e : new Error('Error during copying');
			return false;
		}
	}

	async function read(): Promise<string | undefined> {
		if (!isSupported) {
			error = new Error('Clipboard API not supported');
			console.warn(error);
			return;
		}

		try {
			const clipboardText = await navigator.clipboard.readText();
			text = clipboardText;
			error = null;
			return clipboardText;
		} catch (e) {
			console.error(e);
			error = e instanceof Error ? e : new Error('Error during reading');
			return;
		}
	}

	return {
		get text() {
			return text;
		},
		get error() {
			return error;
		},
		get isSupported() {
			return isSupported;
		},
		copy,
		read
	};
}
