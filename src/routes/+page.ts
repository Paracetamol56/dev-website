import axios from "axios"
import type { PageLoad } from "./$types";

const tools = [
  {
    title: 'Word cloud',
    description: 'Interractive word cloud to know public opinion about a subject',
    path: 'word-cloud',
  },
  {
    title: 'Ormi',
    description: 'ToDo app to encourage productivity take advantage of your momentum',
    path: 'ormi',
  },
]

export const load: PageLoad = async () => {
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