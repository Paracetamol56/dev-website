import db from '$lib/db';
import { error, json, type RequestHandler } from '@sveltejs/kit';
import { ObjectId } from 'mongodb';
import { variants } from '@catppuccin/palette';
import { readToken } from '$lib/token';

export const GET: RequestHandler = async ({ params, cookies }) => {
	const decoded = readToken(cookies);
	const { id } = params;

	const user = await db.collection('users').findOne({ _id: new ObjectId(id), deletedAt: { $exists: false } });
	if (user === null) {
		throw error(404, 'User not found');
	}

	if (user._id !== decoded.userId) {
		// Return the full user object
		return json(user);
	}
	// Return a partial user object
	return json({
		_id: user._id,
		name: user.name
	});
};

export const PATCH: RequestHandler = async ({ params, request, cookies }) => {
	const decoded = readToken(cookies);
	const { id } = params;

	const user = await db.collection('users').findOne({ _id: new ObjectId(id), deletedAt: { $exists: false } });
	if (user === null) {
		throw error(404, 'User not found');
	}

	if (!user._id.equals(decoded.userId)) {
		throw error(403, 'Forbidden');
	}

	const { name, flavour, profilePicture } = await request.json();

	// Validation
	if (name) {
		if (typeof name !== 'string') {
			throw error(400, 'Invalid name');
		}
		if (name.length < 3) {
			throw error(400, 'Name too short');
		}
		if (name.length > 100) {
			throw error(400, 'Name too long');
		}
	}
	/*if (email) {
    if (typeof email !== "string") {
      throw error(400, "Invalid email");
    }
    if (!/\S+@\S+\.\S+/.test(email)) {
      throw error(400, "Invalid email");
    }
  }*/
	if (flavour) {
		if (typeof flavour !== 'string') {
			throw error(400, 'Invalid flavour');
		}
		if (!Object.keys(variants).includes(flavour)) {
			throw error(400, 'Invalid flavour');
		}
	}
	if (profilePicture) {
		if (typeof profilePicture !== 'string') {
			throw error(400, 'Invalid profile picture');
		}
		try {
			new URL(profilePicture);
		} catch (e) {
			throw error(400, 'Invalid profile picture');
		}
	}

	const result = await db.collection('users').updateOne(
		{ _id: new ObjectId(id) },
		{
			$set: {
				name: name ?? user.name,
				flavour: flavour ?? user.flavour,
				profilePicture: profilePicture ?? user.profilePicture
			}
		}
	);

	if (result.acknowledged !== true) {
		throw error(500, 'Failed to update user');
	}

	return new Response(null, { status: 204 });
};

export const DELETE: RequestHandler = async ({ params, cookies }) => {
	const decoded = readToken(cookies);
	const { id } = params;

	const user = await db.collection('users').findOne({ _id: new ObjectId(id), deletedAt: { $exists: false } });
	if (user === null) {
		throw error(404, 'User not found');
	}

	if (!user._id.equals(decoded.userId)) {
		throw error(403, 'Forbidden');
	}

	const result = await db.collection('users').updateOne(
		{ _id: new ObjectId(id) },
		{
			$set: {
				deletedAt: new Date(),
			}
		}
	);

	if (result.acknowledged !== true) {
		throw error(500, 'Failed to delete user');
	}

	return new Response(null, { status: 204 });
};
