import db from "$lib/db";

export const GET = async () => {
  const todos = await db.collection("ormi").find().toArray();
  return new Response(JSON.stringify(todos));
}