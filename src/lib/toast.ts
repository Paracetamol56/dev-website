import { createToaster } from '@melt-ui/svelte';
import type { ToastData } from './toastData';

export const {
	elements,
	helpers: { addToast },
	states: { toasts },
	actions: { portal }
} = createToaster<ToastData>();
