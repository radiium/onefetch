export type AsyncStateOptions<T = unknown> = {
	initial?: T | null | undefined;
	immediate?: boolean;
	resetOnExecute?: boolean;
	onSuccess?: (data: T) => void;
	onError?: (error: unknown) => void;
};

export type AsyncStateReturn<T, TParams = unknown> = {
	current: T | null | undefined;
	readonly loading: boolean;
	readonly error: Error | null;
	execute: (params?: TParams) => Promise<void>;
	reset: () => void;
};

export function useAsyncState<T = unknown, TParams = unknown>(
	promiseFn: (params?: TParams) => Promise<T>,
	options: AsyncStateOptions<T> = {}
): AsyncStateReturn<T, TParams> {
	const {
		initial = null,
		immediate = true,
		resetOnExecute = false,
		onSuccess = () => {},
		onError = () => {}
	} = options;

	let current = $state<T | null | undefined>(initial);
	let loading = $state<boolean>(false);
	let error = $state<Error | null>(null);

	async function execute(params?: TParams): Promise<void> {
		if (resetOnExecute) {
			current = initial;
		}

		loading = true;
		error = null;

		try {
			current = await promiseFn(params);
			onSuccess(current);
		} catch (e) {
			error = e as Error;
			onError(e);
		} finally {
			loading = false;
		}
	}

	function reset() {
		current = initial;
		loading = false;
		error = null;
	}

	if (immediate) {
		execute();
	}

	return {
		get current() {
			return current as T;
		},
		set current(data: T) {
			current = data;
		},
		get loading() {
			return loading;
		},
		get error() {
			return error;
		},
		execute,
		reset
	};
}
