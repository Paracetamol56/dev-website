import type { Place } from '$lib/place';

// ERA5 daily values come from Open-Meteo's historical weather API, which serves the
// Copernicus reanalysis without needing a CDS account.

export type Dataset = 'era5' | 'era5_land';
export type Variable = 'temperature_2m_mean' | 'temperature_2m_max' | 'temperature_2m_min';
export type Resolution = 'month' | 'week' | 'day';

export const DATASETS: Record<Dataset, { label: string; start: number }> = {
	era5: { label: 'ERA5, 0.25° grid (~28 km), from 1940', start: 1940 },
	era5_land: { label: 'ERA5-Land, 0.1° grid (~9 km), land only, from 1950', start: 1950 }
};

export const VARIABLES: Record<Variable, string> = {
	temperature_2m_mean: 'Daily mean',
	temperature_2m_max: 'Daily maximum',
	temperature_2m_min: 'Daily minimum'
};

/** Palette colours the year lines run through, oldest first */
export const YEAR_COLORS = ['blue', 'teal', 'green', 'yellow'] as const;

export type Daily = {
	/** Centre of the grid cell the values belong to */
	lat: number;
	lon: number;
	elevation: number;
	years: number[];
	months: number[];
	days: number[];
	values: (number | null)[];
};

type Download = {
	lat: number;
	lon: number;
	elevation: number;
	dates: string[];
	values: (number | null)[];
};

/** A download kept in the browser, with the day it was last brought up to date */
type Stored = Download & { checked: string };

const STORE = 'era5-v1';
const memory = new Map<string, Promise<Daily>>();

async function readStored(key: string): Promise<Stored | null> {
	try {
		const response = await (await caches.open(STORE)).match(key);
		return response ? await response.json() : null;
	} catch {
		return null;
	}
}

async function writeStored(key: string, stored: Stored) {
	try {
		await (await caches.open(STORE)).put(key, new Response(JSON.stringify(stored)));
	} catch {
		// Without Cache Storage (private windows, full quota) the data is just downloaded again
	}
}

async function download(
	place: Place,
	dataset: Dataset,
	variable: Variable,
	start: string,
	end: string
): Promise<Download> {
	// One variable per request keeps the long range within the API's per-minute allowance
	const params = new URLSearchParams({
		latitude: String(place.lat),
		longitude: String(place.lon),
		start_date: start,
		end_date: end,
		daily: variable,
		models: dataset,
		timezone: 'auto'
	});
	const response = await fetch(`https://archive-api.open-meteo.com/v1/archive?${params}`);
	const body = await response.json().catch(() => null);
	if (!response.ok) throw new Error(body?.reason ?? `ERA5 request failed (${response.status}).`);
	return {
		lat: body.latitude,
		lon: body.longitude,
		elevation: body.elevation,
		dates: body.daily.time,
		values: body.daily[variable]
	};
}

const nextDay = (date: string) =>
	new Date(Date.parse(`${date}T00:00:00Z`) + 86400000).toISOString().slice(0, 10);

async function loadStored(
	key: string,
	place: Place,
	dataset: Dataset,
	variable: Variable
): Promise<Download> {
	const today = new Date().toISOString().slice(0, 10);
	const stored = await readStored(key);
	if (stored?.checked === today) return stored;

	if (!stored) {
		const fresh = await download(
			place,
			dataset,
			variable,
			`${DATASETS[dataset].start}-01-01`,
			today
		);
		if (!fresh.values.some((value) => value !== null)) {
			throw new Error('No ERA5 data here. ERA5-Land has no values over the sea.');
		}
		await writeStored(key, { ...fresh, checked: today });
		return fresh;
	}

	// Published ERA5 days never change, so only the days after the last value are asked again
	const kept = stored.values.findLastIndex((value) => value !== null) + 1;
	const from = stored.dates[kept] ?? nextDay(stored.dates[stored.dates.length - 1]);
	try {
		const recent = from <= today ? await download(place, dataset, variable, from, today) : null;
		const updated: Stored = {
			...stored,
			dates: stored.dates.slice(0, kept).concat(recent?.dates ?? stored.dates.slice(kept)),
			values: stored.values.slice(0, kept).concat(recent?.values ?? stored.values.slice(kept)),
			checked: today
		};
		await writeStored(key, updated);
		return updated;
	} catch {
		// Rate limited or offline: the saved years are still worth showing
		return stored;
	}
}

/** Daily values from the browser's cache when possible, downloading only what is missing. */
export function fetchDaily(place: Place, dataset: Dataset, variable: Variable): Promise<Daily> {
	const key = `/era5-cache/${place.lat.toFixed(4)},${place.lon.toFixed(4)}/${dataset}/${variable}`;
	const cached = memory.get(key);
	if (cached) return cached;

	const request = loadStored(key, place, dataset, variable)
		.then(({ lat, lon, elevation, dates, values }) => ({
			lat,
			lon,
			elevation,
			years: dates.map((date) => Number(date.slice(0, 4))),
			months: dates.map((date) => Number(date.slice(5, 7)) - 1),
			days: dates.map((date) => Number(date.slice(8, 10))),
			values
		}))
		.catch((error) => {
			memory.delete(key);
			throw error;
		});
	memory.set(key, request);
	return request;
}

const MONTH_START = [0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365];
export const MONTH_NAMES = [
	'Jan',
	'Feb',
	'Mar',
	'Apr',
	'May',
	'Jun',
	'Jul',
	'Aug',
	'Sep',
	'Oct',
	'Nov',
	'Dec'
];

/** Day of a 365-day year (0-based); 29 February shares 28 February's slot */
const dayOfYear = (month: number, day: number) =>
	MONTH_START[month] + Math.min(day, 28 + (month === 1 ? 0 : 3)) - 1;

export type Bin = {
	label: string;
	/** Position around the year, from 0 (1 January) to 1 */
	angle: number;
};

export function bins(resolution: Resolution): Bin[] {
	if (resolution === 'month') {
		return MONTH_NAMES.map((label, m) => ({
			label,
			angle: (MONTH_START[m] + MONTH_START[m + 1]) / 2 / 365
		}));
	}
	const size = resolution === 'week' ? 7 : 1;
	const count = Math.floor(365 / size);
	return Array.from({ length: count }, (_, i) => {
		const start = i * size;
		const month = MONTH_START.findLastIndex((first) => first <= start);
		const label =
			resolution === 'week'
				? `Week ${i + 1} (from ${start - MONTH_START[month] + 1} ${MONTH_NAMES[month]})`
				: `${start - MONTH_START[month] + 1} ${MONTH_NAMES[month]}`;
		return { label, angle: (start + size / 2) / 365 };
	});
}

function binOf(resolution: Resolution, month: number, day: number): number {
	if (resolution === 'month') return month;
	const doy = dayOfYear(month, day);
	return resolution === 'week' ? Math.min(Math.floor(doy / 7), 51) : doy;
}

export type Year = {
	year: number;
	values: (number | null)[];
	/** Mean over the year, when every bin has a value */
	annual: number | null;
};

/** Days on each side of a daily value that are averaged with it, to show the trend through the noise */
const DAILY_SMOOTHING = 7;

function smooth(values: (number | null)[]): (number | null)[] {
	const n = values.length;
	return values.map((value, i) => {
		if (value === null) return null;
		let sum = 0;
		let count = 0;
		for (let k = -DAILY_SMOOTHING; k <= DAILY_SMOOTHING; k++) {
			const other = values[(i + k + n) % n];
			if (other !== null) {
				sum += other;
				count++;
			}
		}
		return sum / count;
	});
}

export type Climate = {
	bins: Bin[];
	years: Year[];
};

export function aggregate(daily: Daily, resolution: Resolution): Climate {
	const binList = bins(resolution);
	const count = binList.length;
	const first = daily.years[0];
	const grouped = Array.from({ length: daily.years.at(-1)! - first + 1 }, () => ({
		sums: new Float64Array(count),
		counts: new Uint32Array(count)
	}));
	daily.values.forEach((value, i) => {
		if (value === null) return;
		const year = grouped[daily.years[i] - first];
		const bin = binOf(resolution, daily.months[i], daily.days[i]);
		year.sums[bin] += value;
		year.counts[bin]++;
	});

	// A bin is left out of a year when it has less than half its days,
	// such as the months of the current year that ERA5 has not reached yet
	const fullest = Array.from({ length: count }, (_, i) =>
		Math.max(...grouped.map(({ counts }) => counts[i]))
	);
	const years = grouped.map(({ sums, counts }, y) => {
		const raw = Array.from(sums, (sum, i) =>
			counts[i] * 2 >= fullest[i] ? sum / counts[i] : null
		);
		const values = resolution === 'day' ? smooth(raw) : raw;
		return {
			year: first + y,
			values,
			annual: values.every((value) => value !== null)
				? (values as number[]).reduce((a, b) => a + b, 0) / values.length
				: null
		};
	});

	return { bins: binList, years };
}
