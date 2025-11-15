export function isValid1FichierUrl(url?: string): boolean {
	// Basic type check
	if (typeof url !== 'string' || url.length === 0) {
		return false;
	}

	// Regular expression pattern to validate the format
	// ^ = beginning of character string
	// https:\/\/1fichier\.com\/\? = exact URL
	// [a-z0-9]+ = primary alphanumeric code (lowercase letters and numbers)
	// (&[^&=]+=[^&=]+)* = Optional additional parameters (&key=value)
	// $ = end of character string
	const pattern = /^https:\/\/1fichier\.com\/\?[a-z0-9]+(&[^&=]+=[^&=]+)*$/;

	return pattern.test(url);
}
