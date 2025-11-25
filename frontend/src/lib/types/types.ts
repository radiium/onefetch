export enum DownloadStatus {
	IDLE = 'IDLE',
	PENDING = 'PENDING',
	REQUESTING_INFOS = 'REQUESTING_INFOS',
	REQUESTING_TOKEN = 'REQUESTING_TOKEN',
	DOWNLOADING = 'DOWNLOADING',
	PAUSED = 'PAUSED',
	CANCELLED = 'CANCELLED',
	FAILED = 'FAILED',
	COMPLETED = 'COMPLETED'
}

export enum DownloadType {
	MOVIE = 'MOVIE',
	SERIE = 'SERIE'
}

export interface Download {
	id: string;
	fileName: string;
	customFileName?: string;
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
	isAchived: boolean;
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
	customFileName?: string;
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
	apiKey1fichier: string;
	apiKeyJellyfin: string;
	downloadPath: string;
}

export interface Fileinfo {
	url: string;
	filename: string;
	size: number;
	checksum: string;
	'content-type': string;
	description: string;
	pass: number;
	path: string;
	folder_id: string;
}

export interface DownloadInfoResponse {
	fileinfo: Fileinfo;
	directories: Record<DownloadType, string[]>;
}

export type FSNode = {
	name: string;
	path: string;
	size: number;
	modTime: Date;
	isDir: boolean;
	isHidden: boolean;
	isTmp: boolean;
	isReadOnly: boolean;
	children: FSNode[];
};
