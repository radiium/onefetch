export enum DownloadStatus {
	PENDING = 'PENDING',
	REQUESTING = 'REQUESTING',
	DOWNLOADING = 'DOWNLOADING',
	PAUSED = 'PAUSED',
	COMPLETED = 'COMPLETED',
	FAILED = 'FAILED',
	CANCELLED = 'CANCELLED'
}

export enum DownloadType {
	MOVIE = 'MOVIE',
	TVSHOW = 'TVSHOW'
}

export interface Download {
	id: string;
	fileName: string;
	fileUrl: string;
	type: DownloadType;
	status: DownloadStatus;
	progress: number;
	downloadedBytes: string;
	fileSize?: string;
	speed?: number;
	errorMessage?: string;
	createdAt: string;
	startedAt?: string;
	UpdatedAt: string;
	completedAt?: string;
}

export interface Pagination {
	page: number;
	limit: number;
	total: number;
	totalPages: number;
}

export interface DownloadPage {
	data: Download[];
	pagination: Pagination;
}

export interface DownloadFilters {
	status?: DownloadStatus[];
	type?: DownloadType[];
	page?: number;
	limit?: number;
}

export interface DownloadProgressEvent {
	downloadId: string;
	fileName: string;
	status: DownloadStatus;
	progress: number;
	downloadedBytes: string;
	fileSize?: string;
	speed?: number;
}

export interface CreateDownloadInput {
	url: string;
	type: DownloadType;
	fileName?: string;
	fileSize?: number;
}

export interface Settings {
	apiKey: string;
	downloadPath: string;
}

export type CallbackEmpty = () => void;
export type CallbackError = (error?: unknown) => void;
export type CallbackProgress = (data: DownloadProgressEvent) => void;
export type Callback = CallbackEmpty | CallbackError | CallbackProgress;
