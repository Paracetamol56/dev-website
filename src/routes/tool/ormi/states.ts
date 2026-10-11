import { Ban, Check, Circle, Pause, Play } from 'lucide-svelte';

export const todoStates = {
	TODO: { label: 'To do', icon: Circle, color: 'text-ctp-subtext0' },
	IN_PROGRESS: { label: 'In progress', icon: Play, color: 'text-ctp-blue' },
	STANDBY: { label: 'Standby', icon: Pause, color: 'text-ctp-yellow' },
	DONE: { label: 'Done', icon: Check, color: 'text-ctp-green' },
	CANCELLED: { label: 'Cancelled', icon: Ban, color: 'text-ctp-red' }
};

export type TodoState = keyof typeof todoStates;

export const openStates: TodoState[] = ['TODO', 'IN_PROGRESS', 'STANDBY'];
export const closedStates: TodoState[] = ['DONE', 'CANCELLED'];
