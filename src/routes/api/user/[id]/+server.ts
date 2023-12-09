import db from '$lib/db';
import jwt from 'jsonwebtoken';
import { error, json, type RequestHandler } from '@sveltejs/kit';
import { ObjectId } from 'mongodb';
import { env } from '$env/dynamic/private';
import type TokenPayload from '$lib/token';
import { variants } from '@catppuccin/palette';

export const GET: RequestHandler = async ({ params, cookies }) => {
	const { token } = JSON.parse(cookies.get('user')!);
	if (token === undefined) {
		throw error(401, 'No token provided');
	}
	const decoded = jwt.verify(token, env.JWT_SECRET) as TokenPayload;
	console.log(decoded);

	const { id } = params;

	const user = await db.collection('users').findOne({ _id: new ObjectId(id) });
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
	const { token } = JSON.parse(cookies.get('user')!);
	if (token === undefined) {
		throw error(401, 'No token provided');
	}

	let decoded: TokenPayload;
	try {
		decoded = jwt.verify(token, env.JWT_SECRET) as TokenPayload;
	} catch (e) {
		throw error(401, 'Invalid token');
	}
	const { id } = params;

	const user = await db.collection('users').findOne({ _id: new ObjectId(id) });
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
