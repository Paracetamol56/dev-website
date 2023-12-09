import db from '$lib/db';
import { json, type RequestHandler } from '@sveltejs/kit';
import type { StarHR } from '../../../content/hertzsprung-russell-diagram/utils';

export const GET: RequestHandler = async () => {
	// Use the B-V_1 index
	const stars: StarHR[] = (await db
		.collection('hipparcos')
		.aggregate([
			{
				$match: {
					'B-V': { $gte: -0.4, $lte: 2.2 },
					Plx: { $gte: 0.1 }
				}
			},
			{
				$project: {
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
			}
		])
		.toArray()) as StarHR[];

	return json(stars);
};
