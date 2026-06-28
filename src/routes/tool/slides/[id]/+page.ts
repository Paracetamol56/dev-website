import api from "$lib/api";
import type SlideDeck from "$lib/models/slideDeck";
import { error } from "@sveltejs/kit";

export async function load({ params }) {
  let slideDeck: SlideDeck | undefined;
  try {
    slideDeck = await api.callWithAuth('GET', `/slides/${params.id}`)
      .then((res) => {
        return res.data as SlideDeck;
      })
  } catch (err: any) {
    if (err.response.status === 404) {
      error(404, 'This slide deck does not exist');
    } else {
      error(500, 'An error occured while fetching the slide deck');
    }
  }

  return {
    slideDeck
  };
};
