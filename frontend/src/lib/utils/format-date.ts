export interface FormatDateOptions {
	locale?: string;
	includeTime?: boolean;
	dateStyle?: 'full' | 'long' | 'medium' | 'short';
	timeStyle?: 'full' | 'long' | 'medium' | 'short';
}

export function formatDate(isoString: string, options: FormatDateOptions = {}): string {
	const {
		locale = 'fr-FR', //
		includeTime = true,
		dateStyle = 'short',
		timeStyle = 'short'
	} = options;

	const date = new Date(isoString);

	if (isNaN(date.getTime())) {
		return 'Date invalide';
	}

	const formatOptions: Intl.DateTimeFormatOptions = {
		dateStyle,
		...(includeTime && { timeStyle })
	};

	return date.toLocaleString(locale, formatOptions);
}
