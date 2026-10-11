// Simulates an audience answering a word cloud session.
//
//   node scripts/word-cloud-audience.js <participants> <code> [--api http://localhost:8080/api] [--fast]
//
// Every run answers the same question, so create the session with it:
const QUESTION = 'What makes a great team?';

const COMMON = [
	'trust',
	'communication',
	'respect',
	'fun',
	'feedback',
	'support',
	'collaboration',
	'listening',
	'honesty',
	'empathy',
	'autonomy',
	'transparency',
	'kindness',
	'teamwork'
];
const CREATIVE = [
	'psychological safety',
	'shared goals',
	'coffee',
	'humour',
	'curiosity',
	'ownership',
	'celebrating wins',
	'diversity',
	'patience',
	'clear roles',
	'mentoring',
	'pizza fridays',
	'rubber ducking',
	'blameless retros',
	'code review',
	'pairing',
	'documentation',
	'courage',
	'growth mindset',
	'flexibility',
	'calm',
	'energy',
	'reliability',
	'good snacks'
];
const FRENCH = ['confiance', 'bienveillance', 'entraide', 'écoute', 'humour', 'génial'];
const GRUMPY = ['meetings', 'deadlines', 'micromanagement', 'stress', 'politics', 'burnout'];

const args = process.argv.slice(2);
const option = (name, fallback) => {
	const index = args.indexOf(name);
	return index >= 0 ? args.splice(index, 2)[1] : fallback;
};
const flag = (name) => {
	const index = args.indexOf(name);
	return index >= 0 && args.splice(index, 1).length > 0;
};
const API = option('--api', process.env.WORD_CLOUD_API ?? 'http://localhost:8080/api');
const FAST = flag('--fast');
const [count, code] = [Number(args[0]), args[1]];
if (!Number.isInteger(count) || count < 1 || !code) {
	console.error(
		'Usage: node scripts/word-cloud-audience.js <participants> <code> [--api URL] [--fast]'
	);
	process.exit(1);
}

const random = (min, max) => min + Math.random() * (max - min);
const chance = (probability) => Math.random() < probability;
const pick = (items) => items[Math.floor(Math.random() * items.length)];
const pickWeighted = (items) => items[Math.floor(Math.random() ** 2 * items.length)];
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, FAST ? ms / 10 : ms));

const NAMES = [
	'Ada',
	'Bo',
	'Chloé',
	'Dev',
	'Eli',
	'Fatou',
	'Gus',
	'Hana',
	'Ivo',
	'Jade',
	'Kai',
	'Lou'
];
function personality(index) {
	const inspired = chance(0.75);
	return {
		name: `${NAMES[index % NAMES.length]}${
			index >= NAMES.length ? Math.floor(index / NAMES.length) : ''
		}`,
		// Seconds between two answers, and before the first one
		pace: random(2, 15),
		hesitation: random(0, inspired ? 10 : 40),
		// Answers this participant wants to give
		budget: inspired ? Math.ceil(random(2, 8)) : Math.ceil(random(0, 2)),
		// Likelihood of repeating what others said rather than thinking alone
		conformity: random(0, 0.45),
		// Likelihood of an original answer rather than an obvious one
		creativity: random(0, 1),
		grumpy: chance(0.1),
		french: chance(0.15),
		sloppy: chance(0.3),
		// Everyone has their own original ideas in mind
		ideas: [...CREATIVE].sort(() => Math.random() - 0.5)
	};
}

// What everyone said so far, most recent last, as the room "hears" it
const heard = [];

function nextWord(person, said) {
	const fresh = (words) => words.filter((word) => !said.has(word));
	const recent = fresh(heard.slice(-8));
	let options;
	// Following needs something to follow: a few answers already in the room
	if (heard.length >= 3 && recent.length && chance(person.conformity)) options = recent;
	else if (person.grumpy && chance(0.5)) options = fresh(GRUMPY);
	else if (person.french && chance(0.6)) options = fresh(FRENCH);
	else if (chance(person.creativity)) options = fresh(person.ideas);
	else options = fresh(COMMON);
	if (!options.length) options = fresh([...COMMON, ...CREATIVE]);
	if (!options.length) return null;
	return options === recent ? pick(options) : pickWeighted(options);
}

// Exercises the server's cleaning: capitals, punctuation, extra spaces
function typed(word, person) {
	if (!person.sloppy) return word;
	const variants = [
		(w) => w[0].toUpperCase() + w.slice(1),
		(w) => `${w}!`,
		(w) => `  ${w.replace(' ', '  ')} `,
		(w) => w.toUpperCase()
	];
	return pick(variants)(word);
}

function clientInfo(person) {
	const mobile = chance(0.6);
	return {
		timezone: person.french
			? 'Europe/Paris'
			: pick(['Europe/London', 'Europe/Berlin', 'America/New_York']),
		languages: person.french ? ['fr-FR', 'en'] : ['en-GB', 'en'],
		platform: mobile ? pick(['Android', 'iOS']) : pick(['Windows', 'macOS', 'Linux']),
		mobile,
		screen: mobile
			? pick(['390x844', '412x915', '360x800'])
			: pick(['1920x1080', '2560x1440', '1440x900']),
		availableArea: mobile ? '390x800' : '1920x1040',
		pixelRatio: mobile ? 3 : pick([1, 2]),
		colorDepth: 24,
		cores: pick([4, 6, 8, 12]),
		memory: pick([4, 8]),
		touchPoints: mobile ? 5 : 0,
		colorScheme: pick(['light', 'dark']),
		reducedMotion: false,
		cookiesEnabled: true
	};
}

async function call(method, path, body) {
	const response = await fetch(API + path, {
		method,
		headers: { 'Content-Type': 'application/json' },
		body: body && JSON.stringify(body)
	});
	return {
		status: response.status,
		data: await response.json().catch(() => null),
		headers: response.headers
	};
}

const stats = { sent: 0, refused: 0 };

async function participate(person, sessionId, limit) {
	const { status, data } = await call('POST', `/word-cloud/${sessionId}/participant`);
	if (status !== 201) return console.log(`${person.name} could not join (${status})`);
	const said = new Set();
	const info = clientInfo(person);
	const budget = limit > 0 ? Math.min(person.budget, limit) : person.budget;

	await sleep(person.hesitation * 1000);
	for (let attempt = 0; said.size < budget && attempt < budget * 2; attempt++) {
		const word = nextWord(person, said);
		if (!word) break;
		const result = await call('POST', `/word-cloud/${sessionId}/word`, {
			text: typed(word, person),
			token: data.token,
			client: info
		});
		if (result.status === 201) {
			said.add(word);
			heard.push(word);
			stats.sent++;
			console.log(`${person.name.padEnd(8)} → ${result.data.text}`);
		} else if (result.status === 429) {
			await sleep(Number(result.headers.get('Retry-After') ?? 1) * 1000);
			continue;
		} else {
			stats.refused++;
			if (result.data?.error === 'word cloud closed')
				return console.log(`${person.name} left: session closed`);
			if (result.data?.error === 'word limit reached') break;
		}
		await sleep(random(0.6, 1.4) * person.pace * 1000);
	}
}

const session = await call('GET', `/word-cloud?code=${encodeURIComponent(code)}`);
if (session.status !== 200) {
	console.error(`No open session with code ${code} (${session.status} from ${API})`);
	process.exit(1);
}
console.log(`“${session.data.name}”: ${count} participants answering “${QUESTION}”`);
if (session.data.name !== QUESTION)
	console.log(`(The session asks something else; answers will still be about “${QUESTION}”.)`);

const audience = Array.from({ length: count }, (_, i) => personality(i));
await Promise.all(
	audience.map((person) =>
		participate(person, session.data.id, session.data.settings?.maxWords ?? 0)
	)
);
console.log(`Done: ${stats.sent} answers sent, ${stats.refused} refused.`);
