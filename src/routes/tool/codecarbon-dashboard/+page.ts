import api from "$lib/api";
import type { PageLoad } from "./$types";
import type { Project } from "./projects";

export const load: PageLoad = async () => {
  let projects: Project[];
  projects = await api
    .callWithAuth('GET', '/codecarbon/projects')
    .then((response) => {
      if (response.status !== 200) {
        throw new Error('Failed to fetch projects');
      }
      // Set the projects store with the fetched data
      const data = response.data.map((project: any) => ({
        ...project,
        createdAt: new Date(project.createdAt),
        closedAt: project.closedAt ? new Date(project.closedAt) : null
      }));
      return data
    })
    .catch((error) => {
      console.error('Error fetching projects:', error);
    });

  return {
    projects
  }
}
