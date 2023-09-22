import db from "$lib/db"
import { JWT_SECRET } from '$env/static/private';
import { error, json, type RequestHandler } from "@sveltejs/kit";
import jwt from 'jsonwebtoken';
import type { ObjectId } from "mongodb";

interface TokenPayload {
  userId: ObjectId;
  iat: number;
  exp: number;
}

export const GET: RequestHandler = async ({ url }) => {
  const token = url.searchParams.get("token");

  if (token === null) {
    throw error(400, "No token provided");
  }

  // Use JWT to verify the token
  const user = jwt.verify(token, JWT_SECRET, async (err: jwt.JsonWebTokenError | null, decoded: TokenPayload | undefined) => {
    if (err) {
      console.error(err);
      throw error(400, "Invalid token");
    }

    // Update the user
    const userId: ObjectId = decoded?.userId as ObjectId;
    
    const result = await db.collection("users").updateOne(
      {
        _id: userId,
      },
      {
        $set: {
          lastLogin: new Date(),
        },
      }
    );
  
    // Handle the result
    if (result.acknowledged !== true) {
      console.error(result);
      throw error(500, "Failed to update user");
    }
  
    const user = db.collection("users").findOne({ _id: userId });
    console.log(user);

    // Return the user
    return user;
  });

  return json(
    user,
  );
};