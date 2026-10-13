import { TextColor, TextSize, TextWeight } from './typography';

export enum Tone {
	Accent = 'accent',
	Neutral = 'neutral',
	Success = 'success',
	Warning = 'warning',
	Danger = 'danger'
}

export enum ButtonVariant {
	Solid = 'solid',
	Soft = 'soft',
	Ghost = 'ghost'
}

export const toneText: Record<Tone, TextColor> = {
	[Tone.Accent]: TextColor.Accent,
	[Tone.Neutral]: TextColor.Default,
	[Tone.Success]: TextColor.Success,
	[Tone.Warning]: TextColor.Warning,
	[Tone.Danger]: TextColor.Danger
};

export const toneSolid: Record<Tone, string> = {
	[Tone.Accent]: `bg-ctp-mauve ${TextColor.Inverse}`,
	[Tone.Neutral]: `bg-ctp-surface0 ${TextColor.Default}`,
	[Tone.Success]: `bg-ctp-green ${TextColor.Inverse}`,
	[Tone.Warning]: `bg-ctp-peach ${TextColor.Inverse}`,
	[Tone.Danger]: `bg-ctp-red ${TextColor.Inverse}`
};

export const toneSoft: Record<Tone, string> = {
	[Tone.Accent]: `bg-ctp-mauve/20 ${TextColor.Accent}`,
	[Tone.Neutral]: `bg-ctp-surface0 ${TextColor.Default}`,
	[Tone.Success]: `bg-ctp-green/20 ${TextColor.Success}`,
	[Tone.Warning]: `bg-ctp-peach/20 ${TextColor.Warning}`,
	[Tone.Danger]: `bg-ctp-red/20 ${TextColor.Danger}`
};

export const labelClass = `block ${TextSize.Sm} ${TextWeight.Semibold} ${TextColor.Default}`;

export const controlClass = `w-full min-w-8 rounded-md bg-ctp-surface0 px-3 shadow-md shadow-ctp-crust
	${TextSize.Base} ${TextColor.Default} placeholder:text-ctp-overlay1
	focus:outline-none focus:ring-2 focus:ring-ctp-mauve
	disabled:cursor-not-allowed disabled:opacity-50`;

export const invalidClass = 'ring-2 ring-ctp-red';

let count = 0;
export const uid = (prefix: string) => `${prefix}-${++count}`;
