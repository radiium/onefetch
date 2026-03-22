import { DownloadType } from '$lib/types/types';
import type { IconComponentProps } from 'phosphor-svelte';
import FilmSlateIcon from 'phosphor-svelte/lib/FilmSlateIcon';
import TelevisionIcon from 'phosphor-svelte/lib/TelevisionIcon';
import type { Component } from 'svelte';

export const typeIcons: Record<DownloadType, Component<IconComponentProps>> = {
	[DownloadType.MOVIE]: FilmSlateIcon,
	[DownloadType.SERIE]: TelevisionIcon
};
