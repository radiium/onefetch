type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
type Params = Record<string, string | number | boolean>;

interface RequestConfig<TParams = Params> extends RequestInit {
	params?: TParams;
}

class HttpError extends Error {
	constructor(
		public status: number,
		public statusText: string,
		public data?: unknown
	) {
		super(`HTTP ${status}: ${statusText}`);
		this.name = 'HttpError';
	}

	isClientError(): boolean {
		return this.status >= 400 && this.status < 500;
	}

	isServerError(): boolean {
		return this.status >= 500;
	}
}

class NetworkError extends Error {
	constructor(
		message: string,
		public originalError?: Error
	) {
		super(message);
		this.name = 'NetworkError';
	}
}

export class HttpClient {
	private baseURL: string;
	private defaultHeaders: HeadersInit;

	constructor(baseURL = '', defaultHeaders: HeadersInit = {}) {
		this.baseURL = baseURL;
		this.defaultHeaders = {
			'Content-Type': 'application/json',
			...defaultHeaders
		};
	}

	private buildURL<TParams>(endpoint: string, params?: TParams): string {
		const url = new URL(endpoint, this.baseURL);

		if (params) {
			Object.entries(params).forEach(([key, value]) => {
				if (value !== null && value !== undefined) {
					url.searchParams.append(key, String(value));
				}
			});
		}

		return url.toString();
	}

	private async request<T, TParams>(
		method: HttpMethod,
		endpoint: string,
		config: RequestConfig<TParams> = {}
	): Promise<T> {
		const { params, headers, body, ...restConfig } = config;

		const url = this.buildURL(endpoint, params);

		try {
			const response = await fetch(url, {
				method,
				headers: { ...this.defaultHeaders, ...headers },
				body: body ? JSON.stringify(body) : null,
				...restConfig
			});

			if (!response.ok) {
				const errorData = await this.parseResponse(response);
				throw new HttpError(response.status, response.statusText, errorData);
			}

			return await this.parseResponse<T>(response);
		} catch (error) {
			if (error instanceof HttpError) {
				throw error;
			}

			if (error instanceof TypeError) {
				throw new NetworkError('Network connection failed', error);
			}

			if (error instanceof Error && error.name === 'AbortError') {
				throw new NetworkError('Request canceled', error);
			}

			throw error;
		}
	}

	private async parseResponse<T>(response: Response): Promise<T> {
		const contentType = response.headers.get('content-type');

		if (contentType?.includes('application/json')) {
			return await response.json();
		}

		return (await response.text()) as T;
	}

	async get<T, TParams = Params>(endpoint: string, config?: RequestConfig<TParams>): Promise<T> {
		return this.request<T, TParams>('GET', endpoint, config);
	}

	async post<T, TParams = Params>(
		endpoint: string,
		body?: unknown,
		config?: RequestConfig<TParams>
	): Promise<T> {
		return this.request<T, TParams>('POST', endpoint, { ...config, body: body as BodyInit });
	}

	async put<T, TParams = Params>(
		endpoint: string,
		body?: unknown,
		config?: RequestConfig<TParams>
	): Promise<T> {
		return this.request<T, TParams>('PUT', endpoint, { ...config, body: body as BodyInit });
	}

	async patch<T, TParams = Params>(
		endpoint: string,
		body?: unknown,
		config?: RequestConfig<TParams>
	): Promise<T> {
		return this.request<T, TParams>('PATCH', endpoint, { ...config, body: body as BodyInit });
	}

	async delete<T, TParams = Params>(endpoint: string, config?: RequestConfig<TParams>): Promise<T> {
		return this.request<T, TParams>('DELETE', endpoint, config);
	}
}
