import { env } from '$env/dynamic/private';
import { error } from '@sveltejs/kit';
import type { ObjectId } from 'mongodb';
import jwt from 'jsonwebtoken';

export default interface TokenPayload {
	userId: ObjectId;
	iat: number;
	exp: number;
}

export const readToken = (cookies: any) => {
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

	return decoded;
};
