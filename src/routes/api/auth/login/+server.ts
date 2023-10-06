import db from "$lib/db"
import { JWT_SECRET, SENDGRID_API_KEY } from '$env/static/private';
import { error, json, type RequestHandler } from '@sveltejs/kit';
import jwt from 'jsonwebtoken';
import sgMail from '@sendgrid/mail';
import type { Document, WithId } from "mongodb";

export const POST: RequestHandler = async ({ request }) => {
  const { email } = await request.json();

  let user: WithId<Document> | null = await db.collection("users").findOne({ email: email });
  if (user === null) {
    // Insert the user into the database
    const result = await db.collection("users").insertOne(
      {
        email: email,
        flavour: "mocha",
        createdAt: new Date(),
      }
    );

    // Handle the result
    if (result.acknowledged !== true) {
      console.error(result);
      throw error(500, "Failed to create user");
    }
    user = await db.collection("users").findOne({ email: email });
  }

  // Generate a token
  const link = `${request.headers.get('origin')}/verify?token=${generateToken(user?._id.toHexString() as string)}&redirect=${request.headers.get('referer')}`;
  
  // Send an email
  let mailId: string;
  try {
    const mail: [sgMail.ClientResponse, object] = await sendEmail(email, link);
    mailId = mail[0].headers['x-message-id'];
  } catch (e) {
    console.error(e);
    throw error(500, "Failed to send email");
  }

  return json({
    email: email,
    userId: user?._id.toHexString(),
    mailId: mailId,
  });
};

const generateToken = (userId: string) => {
  // Use JWT to generate a token
  return jwt.sign(
    {
      userId: userId,
    },
    JWT_SECRET,
    {
      expiresIn: "1h",
    }
  );
};

const sendEmail = (email: string, link: string) => {
  sgMail.setApiKey(SENDGRID_API_KEY)
  const msg: sgMail.MailDataRequired = {
    to: email, // Change to your recipient
    from: 'matheo.galu56@gmail.com', // Change to your verified sender
    subject: 'Sign in to dev.matheo-galuba.com',
    content: [{
      type: 'text/plain',
      value: `Click here to sign in: ${link}`,
    }],
  };
  return sgMail.send(msg);
};
