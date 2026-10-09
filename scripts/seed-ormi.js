// Usage: docker exec -i -e SEED_EMAIL=you@example.com dev-db-1 \
//   mongosh --quiet -u root -p example --authenticationDatabase admin dev --file /dev/stdin < scripts/seed-ormi.js
// Re-running replaces the previously seeded todos and leaves the real ones untouched.

const email = process.env.SEED_EMAIL;
const user = db.users.findOne({ email });
if (!user) {
	print(`No user with email ${email}`);
	quit(1);
}

const DAY = 24 * 60 * 60 * 1000;
const now = Date.now();
const random = (max) => Math.floor(Math.random() * max);
const pick = (values) => values[random(values.length)];

const verbs = ['Fix', 'Write', 'Review', 'Refactor', 'Plan', 'Research', 'Update', 'Clean up', 'Test', 'Ship'];
const subjects = [
	'the login flow',
	'the activity calendar',
	'the blog post draft',
	'the icon tile generator',
	'the deployment pipeline',
	'the word cloud tool',
	'the tax paperwork',
	'the reading list',
	'the API documentation',
	'the homepage layout',
	'the gym schedule',
	'the dependency upgrades'
];
const labels = ['home', 'work', 'urgent', 'ormi', 'dev-website', 'reading', 'health', 'finance', 'bug', 'idea', 'chore', 'learning'];

const makeTodo = (createdAt) => ({
	user: user._id,
	title: `${pick(verbs)} ${pick(subjects)}`,
	description: Math.random() < 0.3 ? 'Seeded todo, feel free to delete.' : '',
	labels: [...new Set(Array.from({ length: random(3) }, () => pick(labels)))],
	history: [{ previousState: '', newState: 'TODO', updatedAt: new Date(createdAt) }],
	state: 'TODO',
	position: 0,
	createdAt: new Date(createdAt),
	seeded: true
});

const moveTo = (todo, state, at) => {
	todo.history.push({ previousState: todo.state, newState: state, updatedAt: new Date(at) });
	todo.state = state;
};

const todos = [];

for (let daysAgo = 364; daysAgo >= 0; daysAgo--) {
	const weekend = [0, 6].includes(new Date(now - daysAgo * DAY).getDay());
	const active = daysAgo < 9 || Math.random() < (weekend ? 0.15 : 0.4);
	if (!active) continue;

	const closedAt = daysAgo === 0 ? now - random(60) * 60 * 1000 : now - daysAgo * DAY + random(8 * 60) * 60 * 1000;
	for (let i = 0; i < 1 + random(3); i++) {
		const todo = makeTodo(closedAt - (1 + random(20)) * DAY);
		const createdAt = todo.createdAt.getTime();
		if (Math.random() < 0.6) moveTo(todo, 'IN_PROGRESS', createdAt + (closedAt - createdAt) * 0.4);
		if (todo.state === 'IN_PROGRESS' && Math.random() < 0.2) {
			moveTo(todo, 'STANDBY', createdAt + (closedAt - createdAt) * 0.6);
			moveTo(todo, 'IN_PROGRESS', createdAt + (closedAt - createdAt) * 0.8);
		}
		moveTo(todo, Math.random() < 0.07 ? 'CANCELLED' : 'DONE', closedAt - i * 60 * 1000);
		if (Math.random() < 0.4) todo.dueDate = new Date(closedAt + (random(7) - 3) * DAY);
		todos.push(todo);
	}
}

let position = db.ormi.countDocuments({ user: user._id, seeded: { $ne: true } });
for (const state of ['IN_PROGRESS', 'IN_PROGRESS', 'IN_PROGRESS', 'STANDBY', 'STANDBY', ...Array(9).fill('TODO')]) {
	const todo = makeTodo(now - (1 + random(30)) * DAY);
	const createdAt = todo.createdAt.getTime();
	if (state !== 'TODO') moveTo(todo, 'IN_PROGRESS', createdAt + (now - createdAt) * 0.5);
	if (state === 'STANDBY') moveTo(todo, 'STANDBY', createdAt + (now - createdAt) * 0.8);
	if (Math.random() < 0.7) todo.dueDate = new Date(now + (random(14) - 4) * DAY);
	todo.position = position++;
	todos.push(todo);
}

const removed = db.ormi.deleteMany({ user: user._id, seeded: true }).deletedCount;
db.ormi.insertMany(todos);
print(`Removed ${removed} previously seeded todos, inserted ${todos.length} for ${email}`);
