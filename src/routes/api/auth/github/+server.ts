import { env } from '$env/dynamic/private';
import db from '$lib/db.js';
import { error } from '@sveltejs/kit';
import axios from 'axios';

export const GET = async ({ url }) => {
  const code = url.searchParams.get("code");
  if (!code) {
    throw error(400, "No code provided");
  }
  const path: string = url.searchParams.get("path") || "/";

  const githubToken = await axios
    .post(
      `https://github.com/login/oauth/access_token?client_id=${env.GITHUB_CLIENT_ID}&client_secret=${env.GITHUB_CLIENT_SECRET}&code=${code}`,
      {},
      {
        headers: {
          Accept: "application/json",
        },
      }
    )
    .then((response) => {
      return response.data;
    })
    .catch((error) => {
      console.error(error);
      throw error(500, "Failed to get GitHub token");
    });
  
  const user = await axios
    .get("https://api.github.com/user", {
      headers: {
        Authorization: `token ${githubToken.access_token}`,
      },
    })
    .then((response) => {
      return response.data;
    })
    .catch((error) => {
      console.error(error);
      throw error(500, "Failed to get GitHub user");
    });

  const dbUser = await db.collection("users").findOne({ email: user.email });

  if (dbUser) {
    await db.collection("users").updateOne(
      { email: user.email },
      {
        $set: {
          github: user,
          lastLogin: new Date(),
        },
      }
    );
  } else {
    await db.collection("users").insertOne({
      name: user.name || user.login,
      email: user.email,
      flavour: "mocha",
      github: user,
      createdAt: new Date(),
      lastLogin: new Date(),
    });
  }

  return new Response(null, { status: 302, headers: { Location: path } });
};