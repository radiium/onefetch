import { DownloadType } from '$lib/types/types';
import type { IconComponentProps } from 'phosphor-svelte';
import FilmSlate from 'phosphor-svelte/lib/FilmSlate';
import Television from 'phosphor-svelte/lib/Television';
import type { Component } from 'svelte';

export const typeIcons: Record<DownloadType, Component<IconComponentProps>> = {
	[DownloadType.MOVIE]: FilmSlate,
	[DownloadType.SERIE]: Television
};
