import db from "$lib/db";
import jwt from 'jsonwebtoken';
import { error, json, type RequestHandler } from "@sveltejs/kit";
import { JWT_SECRET } from "$env/static/private";

export const GET: RequestHandler = async ({ url }) => {
  // Get the search params from the URL
  const code: string | null = url.searchParams.get("code");

  // If the code is null, throw an error
  if (code === null) {
    throw error(400, "No code provided");
  }

  // Fetch the id of the session in the database
  const res = await db.collection("word_cloud_sessions").findOne(
    {
      code: code,
      open: true,
    }
  );

  // If the session is null, throw an error
  if (res === null) {
    throw error(404, "Session not found");
  }

  // Return the session
  return json({
    id: res._id,
    name: res.name,
    description: res.description,
    code: res.code,
    open: res.open,
  });
};

interface TokenPayload {
  userId: string;
  iat: number;
  exp: number;
}

export const POST: RequestHandler = async ({ request, cookies }) => {
  // get the token from the cookies and validate it
  const { token } = JSON.parse(cookies.get("user")!);
  // Verify JWT
  if (token === undefined) {
    throw error(401, "No token provided");
  }
  const decoded = jwt.verify(token, JWT_SECRET) as TokenPayload;

  const body = await request.json();

  if (!body.name) {
    throw error(400, "No name provided");
  }
  if (body.name.length < 3 || body.name.length > 100) {
    throw error(400, "Invalid name");
  }

  if (body.description && (body.description.length < 10 || body.description.length > 1000)) {
    throw error(400, "Invalid description");
  }

  // Insert the session into the database
  const result = await db.collection("word_cloud_sessions").insertOne({
    user: decoded.userId,
    name: body.name,
    description: body.description,
    code: Math.random().toString(36).substr(2, 5),
    open: true,
    words: [],
    createdAt: new Date(),
  });

  if (result.acknowledged !== true) {
    throw error(500, "Failed to insert session");
  }

  return json({
    id: result.insertedId,
  });
};
