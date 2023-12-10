import db from '$lib/db';
import { readToken } from '$lib/token';
import { error, json, type RequestHandler } from '@sveltejs/kit';
import { ObjectId } from 'mongodb';

export const GET: RequestHandler = async ({ cookies }) => {
	const decoded = readToken(cookies);

	const user = await db
		.collection('users')
		.findOne({ _id: new ObjectId(decoded.userId) }, { projection: { password: 0 } });

	const wordCloudSessions = await db
		.collection('word_cloud_sessions')
		.find({ user: new ObjectId(decoded.userId) })
		.toArray();
	if (user === null) {
		throw error(404, 'User not found');
	}
	return json({
		user: user,
		wordCloudSessions: wordCloudSessions
	});
};
