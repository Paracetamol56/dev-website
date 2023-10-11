import db from "$lib/db";
import { error, json, type RequestHandler } from "@sveltejs/kit";
import { ObjectId } from "mongodb";
import jwt, { type JwtPayload } from "jsonwebtoken";
import { env } from '$env/dynamic/private';

export const GET: RequestHandler = async ({ params, cookies }) => {
  // Get the session
  const id = params.id;
  const result = await db.collection("word_cloud_sessions").findOne(
    {
      _id: new ObjectId(id),
    }
  );
  if (result === null) {
    throw error(404, "Session not found");
  }

  // Check if the user owns the session
  const userCookie = cookies.get("user");
  if (userCookie) {
    const token = JSON.parse(userCookie).token;
    try {
      const decoded = jwt.verify(token, env.JWT_SECRET) as JwtPayload;
      if (decoded.userId === result.user) {
        return json({
          id: result._id,
          name: result.name,
          description: result.description,
          code: result.code,
          open: result.open,
          words: result.words,
          createdAt: result.createdAt,
          closedAt: result.closedAt,
        });
      } else {
        // Not the owner
        throw error(401, "Unauthorized");
      }
    } catch (err) {
      // Invalid token
      throw error(401, "Unauthorized");
    }
  }

  // Return the session without the code
  return json({
    id: result._id,
    name: result.name,
    description: result.description,
    open: result.open,
  });
};

export const PUT: RequestHandler = async ({ params, request }) => {
  const id = params.id;
  const {text, ip} = await request.json();

  // Check if the session exists and is open
  const result = await db.collection("word_cloud_sessions").findOne(
    {
      _id: new ObjectId(id),
      open: true,
    }
  );
  if (result === null) {
    throw error(404, "Session not found");
  }

  // Validate the body
  if (!text) {
    throw error(400, "No word provided");
  }
  if (text.length > 100) {
    throw error(400, "Invalid word");
  }

  // Add the word to the session
  await db.collection("word_cloud_sessions").updateOne(
    {
      _id: new ObjectId(id),
    },
    {
      $push: {
        words: {
          text: text,
          ip: ip ?? "unknown",
          userAgent: request.headers.get("user-agent") ?? "unknown",
          createdAt: new Date(),
        }
      },
    }
  );

  return new Response(null, { status: 204 });
};

export const DELETE: RequestHandler = async ({ params, cookies }) => {
  // Get the session
  const id = params.id;
  const result = await db.collection("word_cloud_sessions").findOne(
    {
      _id: new ObjectId(id),
    }
  );
  if (result === null) {
    throw error(404, "Session not found");
  }

  // Check if the user owns the session
  const userCookie = cookies.get("user");
  if (userCookie) {
    const token = JSON.parse(userCookie).token;
    try {
      const decoded = jwt.verify(token, env.JWT_SECRET) as JwtPayload;
      if (decoded.userId === result.user) {
        // Delete the session
        await db.collection("word_cloud_sessions").updateOne(
          {
            _id: new ObjectId(id),
          },
          {
            $set: {
              open: false,
              closedAt: new Date(),
            },
          }
        );
        return new Response(null, { status: 204 });
      } else {
        // Not the owner
        throw error(401, "Unauthorized");
      }
    } catch (err) {
      // Invalid token
      throw error(401, "Unauthorized");
    }
  }

  // Not logged in
  throw error(401, "Unauthorized");
};