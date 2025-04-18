import { writable } from 'svelte/store';

export interface PartialProject {
  name: string;
  description: string;
  publicLink?: string;
}

export interface Project extends PartialProject {
  id: string;
  runsCount: number;
  createdAt: Date;
  closedAt: Date | null;
}

export const projectsStore = writable<Project[]>([
  {
    id: '1',
    name: 'Project Alpha',
    description: 'Analysis of carbon emissions in transportation sector',
    runsCount: 15,
    publicLink: 'https://example.com/alpha',
    createdAt: new Date(),
    closedAt: null
  },
  {
    id: '2',
    name: 'Project Beta',
    description: 'Machine learning model for carbon footprint calculation',
    runsCount: 8,
    publicLink: 'https://example.com/beta',
    createdAt: new Date(),
    closedAt: new Date(),
  }
]);

