import db from '$lib/db';
import type { PageServerLoad  } from './$types';
import type { StarHR } from './utils';

export const load: PageServerLoad = async () => {
  // Use the B-V_1 index
	const stars: StarHR[] = await db
    .collection('hipparcos')
    .aggregate([
      { $match: {
          "B-V": { $gte: -0.4, $lte: 2.2 },
          Plx: { $gte: 0.1 }
        }
      },
      { $project: {
          _id: 0,
          HIP: 1,
          Amag: {
            $subtract: [
              '$Vmag',
              {
                $multiply: [
                  5,
                  {
                    $log10: {
                      $divide: [
                        {
                          $divide: [1000, '$Plx']
                        },
                        10
                      ]
                    }
                  }
                ]
              }
            ]
          },
          BV: '$B-V' 
        }
      },
    ])
    .toArray() as StarHR[];

  return { stars };
};
