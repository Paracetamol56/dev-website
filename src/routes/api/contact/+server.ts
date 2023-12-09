import { error, type RequestHandler } from '@sveltejs/kit';
import sgMail from '@sendgrid/mail';
import type { ObjectId } from 'mongodb';
import db from '$lib/db';
import { env } from '$env/dynamic/private';

interface Contact {
	userId: ObjectId | null;
	name: string;
	email: string;
	message: string;
}

export const POST: RequestHandler = async ({ request }) => {
	const data: Contact = await request.json();

	if (!data.name) throw error(400, 'Missing field: name');
	if (!data.email) throw error(400, 'Missing field: email');
	if (!data.message) throw error(400, 'Missing field: message');
	if (data.name.length < 3) throw error(400, 'Invalid name');
	if (data.name.length > 100) throw error(400, 'Invalid name');
	if (!/\S+@\S+\.\S+/.test(data.email)) throw error(400, 'Invalid email');
	if (data.message.length < 10) throw error(400, 'Invalid message');
	if (data.message.length > 1000) throw error(400, 'Invalid message');

	let mailId: string;
	try {
		sgMail.setApiKey(env.SENDGRID_API_KEY!);
		const msg: sgMail.MailDataRequired = {
			to: env.ADMIN_EMAIL,
			from: 'matheo.galu56@gmail.com',
			subject: 'New message from dev.matheo-galuba.com',
			content: [
				{
					type: 'text/plain',
					value: `
					User: ${data.userId}
					Name: ${data.name}
					Email: ${data.email}
					Message: ${data.message}
				`
				}
			]
		};
		const sendResult = await sgMail.send(msg);
		mailId = sendResult[0].headers['x-message-id'];
	} catch (e) {
		console.error(e);
		throw error(500, 'Failed to send email');
	}

	db.collection('contacts').insertOne({
		user: data.userId,
		name: data.name,
		email: data.email,
		message: data.message,
		mailId: mailId,
		createdAt: new Date()
	});

	return new Response('OK');
};
