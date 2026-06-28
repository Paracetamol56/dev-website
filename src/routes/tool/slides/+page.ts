import api, { type PaginatedArrayResult } from "$lib/api";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import type SlideDeck from "$lib/models/slideDeck";
import { get } from "svelte/store";
import { user } from "$lib/store";

export async function load({ url }) {
  if (get(user).id == null) {
    throw error(401, 'Unauthorized access. Please log in to view this page.');
  }

  const limit = Number(url.searchParams.get('limit')) || 12;
  const page = Number(url.searchParams.get('page')) || 1;

  if (limit < 1)
    throw error(400, 'Invalid limit value. Must be greater than 0.');
  if (limit > 100)
    throw error(400, 'Invalid limit value. Must be less than or equal to 100.');
  if (page < 1)
    throw error(400, 'Invalid page value. Must be greater than 0.');

  const result: PaginatedArrayResult<SlideDeck> | null = await api
    .callWithAuth('GET', `/slides?limit=${limit}&page=${page}`)
    .then((response) => {
      if (response.status !== 200) {
        throw new Error('Failed to fetch projects');
      }
      return response.data as PaginatedArrayResult<SlideDeck>;
    })
    .catch((error) => {
      console.error('Error fetching projects:', error);
      return null;
    });

  return {
    slides: result,
  };
}

