// A route hides parts of the site's frame by returning { shell: { navbar: false } } from a load function
export type Shell = { navbar: boolean; footer: boolean };

export const defaultShell: Shell = { navbar: true, footer: true };
