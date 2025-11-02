import { browser } from '$app/environment';
import type {
	CreateDownloadInput,
	Download,
	DownloadFilters,
	DownloadPage,
	FileinfoResponse,
	Settings
} from '../types/types';
import { HttpClient } from './http-client';
import { useSSE, type UseSSEOptions, type UseSSEReturn } from './use-sse.svelte';

export const baseUrl = browser ? `${location.protocol}//${location.host}` : '';
export const http = new HttpClient(baseUrl);

export const api = {
	settings: {
		async get(): Promise<Settings> {
			return http.get('/api/settings');
		},
		async update(settings?: Partial<Settings>): Promise<Settings> {
			return http.patch('/api/settings', settings);
		}
	},
	fileinfo: {
		async get(url?: string): Promise<FileinfoResponse> {
			return http.get('/api/fileinfo', { params: { url } });
		}
	},
	download: {
		async getAll(params?: DownloadFilters): Promise<DownloadPage> {
			return http.get('/api/downloads', { params });
		},
		async get(id: string): Promise<Download> {
			return http.get(`/api/downloads/${id}`);
		},
		async create(input?: CreateDownloadInput): Promise<Download> {
			return http.post(`/api/downloads`, input);
		},
		async pause(id?: string): Promise<Download> {
			return http.post(`/api/downloads/${id}/pause`, null);
		},
		async resume(id?: string): Promise<Download> {
			return http.post(`/api/downloads/${id}/resume`, null);
		},
		async cancel(id?: string): Promise<Download> {
			return http.post(`/api/downloads/${id}/cancel`, null);
		},
		async archive(id?: string): Promise<Download> {
			return http.post(`/api/downloads/${id}/archive`, null);
		},
		streams(options: Omit<UseSSEOptions<string>, 'event'>): UseSSEReturn {
			return useSSE(`${baseUrl}/api/downloads/streams`, {
				event: 'progress',
				...options
			});
		}
	}
};
