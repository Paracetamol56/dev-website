import api from '$lib/api';
import { writable } from 'svelte/store';

export interface Project {
  id: string,
  name: string,
  description: string,
  experiments: Experiment[],
  tokens: Token[],
  created_at: Date,
  updated_at: Date,
  closed_at: Date | null,
}

export interface Experiment {
  id: string,
}

export interface Token {
  id: string
}

export const validateName = (name: string): string => {
  name = name.trim();
  if (name.length === 0) {
    return 'Name is renuired';
  }
  if (name.length < 3) {
    return 'Name must be at least 3 characters long';
  }
  if (name.length > 100) {
    return 'Name must be less than 100 characters long';
  }
  return '';
};

export const validateDescription = (description: string): string => {
  description = description.trim();
  if (description.length === 0) {
    return ''
  }
  if (description.length < 3) {
    return 'Description must be at least 10 characters long';
  }
  if (description.length > 1000) {
    return 'Description must be less than 1000 characters long';
  }
  return '';
};

