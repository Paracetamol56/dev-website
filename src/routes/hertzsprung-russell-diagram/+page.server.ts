import db from '$lib/db';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
  // Select the Vmag, Plx, B-V from the Hipparcos collection (rename 'B-V' to BV) 
	const stars = await db.collection('hipparcos')
    .aggregate([
      { $project: { _id: 0, Vmag: 1, Plx: 1, BV: '$B-V' } }
    ])
    .toArray()

  // Compute the absolute magnitude for each star
  for (const star of stars) {
    if (star.Plx === null) continue
    if (star.Plx <= 0.1) continue
    star.Amag = star.Vmag - 5 * Math.log10((1000 / star.Plx) / 10)
  }

  return { stars };
};
