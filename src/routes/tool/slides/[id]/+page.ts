import api from "$lib/api";
import { error } from "@sveltejs/kit";

interface SlideDeck {
  id: string;
  html: string;
  created_at: Date;
  updated_at: Date
}

export const load: PageLoad = async ({ params }) => {
  let deck: SlideDeck | undefined;

  return {
    deck: {
      id: "123",
      html: "<section>\n<h1>Hello world !</h1>\n</section>\n<section>\n<section>\n<h2>A</h2>\n</section>\n<section>\n<h2>B</h2>\n</section>\n</section>",
      created_at: new Date(),
      updated_at: new Date()
    }
  };

  try {
    deck = await api.callWithAuth('GET', `/slides/${params.id}`)
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
    deck
  };
};
