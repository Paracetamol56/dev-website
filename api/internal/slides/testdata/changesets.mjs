// Run from the repository root: bun api/internal/slides/testdata/changesets.mjs > api/internal/slides/testdata/changesets.json
import { ChangeSet, Text } from '@codemirror/state';

let seed = 42;
const random = (n) => {
	seed = (seed + 0x6d2b79f5) | 0;
	let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
	t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
	return Math.floor((((t ^ (t >>> 14)) >>> 0) / 4294967296) * n);
};
const pieces = ['a', 'é', '😀', '\n', '<section>', '</h1>', 'é😀\nx', ' ', '\t', '中文'];
const text = (length) => Array.from({ length }, () => pieces[random(pieces.length)]).join('');

const cases = [];
for (let i = 0; i < 300; i++) {
	const doc = text(random(12));
	const length = doc.length;
	const specs = [];
	for (let n = random(4); n >= 0; n--) {
		const from = random(length + 1);
		const to = Math.min(length, from + random(4));
		specs.push({ from, to, insert: random(3) ? text(random(4)) : '' });
	}
	const changes = ChangeSet.of(specs, length);
	cases.push({ doc, changes: changes.toJSON(), result: changes.apply(Text.of(doc.split('\n'))).toString() });
}
console.log(JSON.stringify(cases, null, '\t'));
