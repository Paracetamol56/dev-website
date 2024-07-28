import { ImageResponse } from '@ethercorps/sveltekit-og';
import { RequestHandler } from '@sveltejs/kit';

export const GET: RequestHandler = async ({ params, fetch }) => {
  const post = await import(`../../../../content/${params.slug}.md`);

  const template = `
  <div
    tw="flex flex-col w-full h-full items-center justify-between px-16 py-8"
    style="background: #1e1e2e; color: #cdd6f4; font-family: 'Inter SemiBold', sans-serif;"
  >
    <div tw="flex-1 flex flex-row w-full items-center justify-between">
      <div tw="flex flex-col items-start justify-center w-2/3">
        <h1 tw="text-6xl text-transparent bg-clip-text leading-1.15" style="background-clip: text; background-image: linear-gradient(to right, rgb(203, 166, 247), rgb(180, 190, 254)); font-family: 'Inter SemiBold', sans-serif;">${post.metadata.title}</h1>
        <h2 tw="text-4xl">${post.metadata.description}</h2>
        <p tw="text-2xl" style="color: #b4befe;">
          ${post.metadata.tags.map((tag: string) => `<span tw="mr-2">#${tag}</span>`).join('')}
        </p>
      </div>
      <img tw="h-32" alt="Logo" src="data:image/png;base64,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">
    </div>

    <p tw="text-2xl ml-auto" style="color: #b4befe;">dev.matheo-galuba.com</p>
  </div>
  `;

  const regularFile = await fetch('/fonts/Inter/static/Inter-Regular.ttf');
  const semiBoldFile = await fetch('/fonts/Inter/static/Inter-SemiBold.ttf');
  const boldFile = await fetch('/fonts/Inter/static/Inter-Bold.ttf');
	const regularData = await regularFile.arrayBuffer();
  const semiBoldData = await semiBoldFile.arrayBuffer();
  const boldData = await boldFile.arrayBuffer();

	return await new ImageResponse(template, {
		height: 630,
		width: 1200,
    fonts: [
      {
        name: 'Inter Regular',
        data: regularData,
        weight: 400,
      },
      {
        name: 'Inter SemiBold',
        data: semiBoldData,
        weight: 600,
      },
      {
        name: 'Inter Bold',
        data: boldData,
        weight: 700,
      },
    ],
	});
};
