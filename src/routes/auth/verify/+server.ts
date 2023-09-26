import db from "$lib/db"
import { JWT_SECRET } from '$env/static/private';
import { error, json, type RequestHandler } from "@sveltejs/kit";
import jwt from 'jsonwebtoken';
import { ObjectId } from "mongodb";

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
  const user = await jwt.verify(token, JWT_SECRET, async (err: jwt.JsonWebTokenError | null, decoded: TokenPayload | undefined) => {
    if (err) {
      console.error(err);
      throw error(400, "Invalid token");
    }

    // Update the user
    const userId: ObjectId = new ObjectId(decoded?.userId);
    console.log(userId);
    
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
    if (result.acknowledged !== true || result.modifiedCount !== 1) {
      console.error(result);
      throw error(500, "Failed to update user");
    }
  
    const user = await db.collection("users").findOne({ _id: userId });

    // Return the user
    return user;
  });

  // Return a redirect with the user
  return json(user);
};