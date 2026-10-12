import type { Component, ComponentType } from 'svelte';
import type { Tone } from './styles';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type IconComponent = ComponentType | Component<any>;

export type MenuItem = {
	label: string;
	icon?: IconComponent;
	tone?: Tone;
	disabled?: boolean;
	onSelect: () => void;
};

export type Section = { id: string; title: string };
