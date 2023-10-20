import type { ObjectId } from "mongodb";

enum TodoState {
  Backlog = 'backlog',
  Pending = 'pending',
  InProgress = 'in-progress',
  Done = 'done',
}

interface Label {
  _id: ObjectId;
  name: string;
  color: string;
  createdAt: Date;
  updatedAt: Date;
}

interface Todo {
  _id?: ObjectId;
  parentId?: ObjectId;
  name: string;
  state: TodoState;
  markdown: string;
  githubUser?: string;
  githubRepo?: string;
  githubIssue?: number;
  links: string[];
  labels: Label[];
  createdAt: Date;
  updatedAt: Date;
  doneAt?: Date;
}

export { type Todo, TodoState, type Label }
