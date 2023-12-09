import db from '$lib/db';
import { env } from '$env/dynamic/private';
import { error, json, type RequestHandler } from '@sveltejs/kit';
import jwt from 'jsonwebtoken';
import { ObjectId } from 'mongodb';
import type TokenPayload from '$lib/token';

export const GET: RequestHandler = async ({ url }) => {
	const token = url.searchParams.get('token');

	if (token === null) {
		throw error(400, 'No token provided');
	}

	// Use JWT to verify the token
	try {
		const decoded = jwt.verify(token, env.JWT_SECRET!) as TokenPayload;

		// Update the user
		const userId: ObjectId = new ObjectId(decoded.userId);

		const result = await db.collection('users').updateOne(
			{
				_id: userId
			},
			{
				$set: {
					lastLogin: new Date()
				}
			}
		);

		// Handle the result
		if (result.acknowledged !== true || result.modifiedCount !== 1) {
			console.error(result);
			throw error(500, 'Failed to update user');
		}

		const user = await db.collection('users').findOne({ _id: userId });

		// Sign a new token
		const newToken = jwt.sign(
			{
				userId: userId
			},
			env.JWT_SECRET,
			{
				expiresIn: '30d'
			}
		);

		// Return the user
		return json({
			token: newToken,
			user: {
				id: user?._id.toHexString(),
				name: user?.name,
				email: user?.email,
				flavour: user?.flavour
			}
		});
	} catch (err) {
		console.error(err);
		throw error(400, 'Invalid token');
	}
};
