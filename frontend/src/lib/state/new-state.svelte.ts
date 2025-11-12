import { goto } from '$app/navigation';
import { api } from '$lib/api/api.svelte';
import { useAsyncState } from '$lib/api/use-async-state.svelte';
import {
	DownloadType,
	type CreateDownloadInput,
	type Download,
	type DownloadInfoResponse,
	type Fileinfo
} from '$lib/types/types';

export type FormState = {
	url: string;
	type: DownloadType;
	fileName: string;
	fileDir: string;
};

function isValid1FichierUrl(url?: string): boolean {
	// Vérification basique du type
	if (typeof url !== 'string' || url.length === 0) {
		return false;
	}

	// Pattern regex pour valider le format
	// ^ = début de chaîne
	// https:\/\/1fichier\.com\/\? = URL exacte
	// [a-z0-9]+ = code alphanumérique (lettres minuscules et chiffres)
	// $ = fin de chaîne
	const pattern = /^https:\/\/1fichier\.com\/\?[a-z0-9]+$/;

	return pattern.test(url);
}

export const createNewState = () => {
	let formState = $state<FormState>({
		url: '',
		type: DownloadType.MOVIE,
		fileName: '',
		fileDir: ''
	});

	let fileinfo = $state<Fileinfo | null>(null);
	let directories = $state<Record<DownloadType, string[]>>({
		[DownloadType.MOVIE]: [],
		[DownloadType.SERIE]: []
	});

	const pathPreview = $derived(
		['downloads', formState.type.toLowerCase() + 's', formState.fileDir.trim(), formState.fileName]
			.filter(Boolean)
			.map((value) => value.trim())
			.join('/')
			.replaceAll('..', '/')
			.replaceAll('//', '/')
	);

	const getState = useAsyncState<DownloadInfoResponse, string>(
		api.download.getInfos, //
		{
			immediate: false,
			onSuccess(value): void {
				fileinfo = value.fileinfo;
				directories = value.directories;
				formState.fileName = fileinfo.filename;
			}
		}
	);

	const createState = useAsyncState<Download, CreateDownloadInput>(
		api.download.create, //
		{
			immediate: false,
			onSuccess() {
				// eslint-disable-next-line svelte/no-navigation-without-resolve
				goto('/download?test=test').then(() => {
					formState = {
						url: '',
						type: DownloadType.MOVIE,
						fileName: '',
						fileDir: ''
					};
				});
			}
		}
	);

	return {
		// Url
		get url() {
			return formState.url;
		},
		set url(value) {
			if (formState.url !== value) {
				formState.url = value;
				if (isValid1FichierUrl(formState.url)) {
					getState.execute(formState.url);
				} else {
					fileinfo = null;
					formState.fileName = '';
				}
			}
		},
		// Type
		get type() {
			return formState.type;
		},
		set type(value) {
			formState.type = value;
		},
		// FileName
		get fileName() {
			return formState.fileName;
		},
		set fileName(value) {
			formState.fileName = value;
		},
		// Filedir
		get fileDir() {
			return formState.fileDir;
		},
		set fileDir(value) {
			formState.fileDir = value;
		},
		// Fileinfo
		get fileinfo() {
			return fileinfo;
		},
		get directories() {
			return directories[formState.type] ?? [];
		},
		// RullPath
		get pathPreview() {
			return pathPreview;
		},
		// States
		get loading() {
			return getState.loading || createState.loading;
		},
		get error() {
			return getState.error || createState.error;
		},
		isValid1FichierUrl,
		create: () => {
			createState.execute($state.snapshot(formState));
		}
	};
};
