import axios from "axios"
import type { PageData } from "./$types";

const tools = [
  {
    title: 'Word cloud',
    description: 'Interractive word cloud to know public opinion about a subject',
    path: 'word-cloud',
  },
  {
    title: 'Word cloud',
    description: 'Interractive word cloud to know public opinion about a subject',
    path: 'word-cloud',
  },
  {
    title: 'Word cloud',
    description: 'Interractive word cloud to know public opinion about a subject',
    path: 'word-cloud',
  },
  {
    title: 'Word cloud',
    description: 'Interractive word cloud to know public opinion about a subject',
    path: 'word-cloud',
  }
]

export const load: PageData = async () => {
  return axios
    .get('/api/content')
    .then((res) => {
      return {pages: res.data, tools}
    })
    .catch((err) => {
      console.error(err)
      return {pages: [], tools}
    });
}