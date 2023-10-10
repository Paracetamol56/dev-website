import axios from "axios"
import type { PageData } from "./$types";

export const load: PageData = async () => {
  return axios
    .get('/api/content')
    .then((res) => {
      return {pages: res.data}
    })
    .catch((err) => {
      console.error(err)
      return {pages: []}
    });
}