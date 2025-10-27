import { DownloadType } from '$lib/types/types';
import FilmSlate from 'phosphor-svelte/lib/FilmSlate';
import Television from 'phosphor-svelte/lib/Television';

export const typeIcons = {
	[DownloadType.MOVIE]: FilmSlate,
	[DownloadType.TVSHOW]: Television
};
