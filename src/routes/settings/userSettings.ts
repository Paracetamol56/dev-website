import type { variants } from "@catppuccin/palette";

export interface UserSettings {
  id: string;
	name: string;
	email: string;
  createdAt: Date;
  lastLogin: Date;
	flavour: keyof typeof variants;
	github: any;
}