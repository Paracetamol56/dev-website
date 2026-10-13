/** Browser details sent with each word, to recognise the same device behind several participants. */
export function clientInfo() {
	const nav = navigator as Navigator & {
		userAgentData?: { platform?: string; mobile?: boolean };
		deviceMemory?: number;
	};
	const size = (width: number, height: number) => `${width}x${height}`;
	return {
		timezone: Intl.DateTimeFormat().resolvedOptions().timeZone ?? '',
		languages: [...(navigator.languages ?? [navigator.language])].slice(0, 10),
		platform: nav.userAgentData?.platform ?? navigator.platform ?? '',
		mobile: nav.userAgentData?.mobile ?? /Mobi/.test(navigator.userAgent),
		screen: size(screen.width, screen.height),
		availableArea: size(screen.availWidth, screen.availHeight),
		pixelRatio: devicePixelRatio,
		colorDepth: screen.colorDepth,
		cores: navigator.hardwareConcurrency ?? 0,
		memory: nav.deviceMemory ?? 0,
		touchPoints: navigator.maxTouchPoints ?? 0,
		colorScheme: matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light',
		reducedMotion: matchMedia('(prefers-reduced-motion: reduce)').matches,
		cookiesEnabled: navigator.cookieEnabled
	};
}
