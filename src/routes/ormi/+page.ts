import axios from "axios"
import type { PageLoad } from "../$types";

export const load: PageLoad = async () => {
  return axios
    .get('/api/ormi')
    .then((res) => {
      return {todos: res.data}
    })
    .catch((err) => {
      console.error(err)
      return {todos: []}
    });
}
