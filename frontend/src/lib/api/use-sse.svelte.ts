export interface UseSSEOptions<T> {
	event?: string;
	onConnect?: () => void;
	onDisconnect?: () => void;
	onError?: (error: Error) => void;
	onMessage?: (event: MessageEvent<T>) => void;
}

export interface UseSSEReturn {
	readonly connected: boolean;
	readonly error: Error | null;
	connect: () => void;
	disconnect: () => void;
}

export function useSSE<T>(url: string, options: UseSSEOptions<T> = {}): UseSSEReturn {
	const {
		event = 'message', //
		onConnect = () => {},
		onDisconnect = () => {},
		onError,
		onMessage
	} = options;

	let connected = $state<boolean>(false);
	let error = $state<Error | null>(null);
	let source = $state<EventSource | null>(null);

	function connect() {
		if (source) return;

		source = new EventSource(url);
		connected = false;
		error = null;

		source.onopen = () => {
			connected = true;
			onConnect();
		};

		source.onerror = () => {
			error = new Error('Erreur de connexion SSE');
			connected = false;
			onError?.(error);
		};

		source.addEventListener(event, (e: MessageEvent<T>) => {
			onMessage?.(e);
		});
	}

	function disconnect() {
		if (source) {
			source.close();
			source = null;
			connected = false;
			onDisconnect();
		}
	}

	return {
		get connected() {
			return connected;
		},
		get error() {
			return error;
		},
		connect,
		disconnect
	};
}
