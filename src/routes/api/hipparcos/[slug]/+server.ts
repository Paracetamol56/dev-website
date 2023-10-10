import db from "$lib/db";
import { error, json, type RequestHandler } from "@sveltejs/kit";
import type { Star, StarHR } from "../../../../content/hertzsprung-russell-diagram/utils";

export const GET: RequestHandler = async ({ params }) => {
  // Validation : slug must be a positive integer
  if (isNaN(parseInt(params.slug!))) {
    throw error(400, "Invalid HIP number"); 
  }

  const star: Star = await db
    .collection("hipparcos")
    .findOne({ HIP: parseInt(params.slug!) }) as Star;
  
  // Validation : star must exist
  if (star === null) {
    throw error(404, "Star not found");
  }

  const starHR: StarHR = {
    HIP: star.HIP,
    Amag: star.Vmag - 5 * Math.log10(1000 / star.Plx) + 5,
    BV: star["B-V"],
  };

  return json(starHR);
};
