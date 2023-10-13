import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { OpenAI } from "openai";

export const POST = async ({ request }) => {
  const data: { prompt: string } = await request.json();
  
  let prompt = data.prompt;
  if (prompt) {
    prompt = `Generate a debate topic about the domain of ${prompt}. Write only the topic.`
  } else {
    prompt = "Generate a debate topic. Write only the topic."
  }

  const openai = new OpenAI({
    organization: "org-V585na3D62todz2USV7x6zr9",
    apiKey: env.OPENAI_API_KEY,
  });

  const chatCompletion = await openai.chat.completions.create({
    messages: [{ role: 'user', content: prompt }],
    model: 'gpt-3.5-turbo',
  });

  return json(chatCompletion);
};
