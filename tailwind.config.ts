import type { Config } from 'tailwindcss';
import catppuccin from "@catppuccin/tailwindcss";
import plugin from 'tailwindcss/plugin';

const config = {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
    container: {
      padding: {
        DEFAULT: '1rem',
        sm: '2rem',
        lg: '3rem',
      },
    },
		extend: {
      heading: {
        '1': {
          fontSize: '5xl',
          fontWeight: 'bold',
          lineHeight: 'tight',
        },
        '2': {
          fontSize: '4xl',
          fontWeight: 'bold',
          lineHeight: 'tight',
        },
        '3': {
          fontSize: '3xl',
          fontWeight: 'bold',
          lineHeight: 'tight',
        },
        '4': {
          fontSize: '2xl',
          fontWeight: 'bold',
          lineHeight: 'tight',
        },
        '5': {
          fontSize: 'xl',
          fontWeight: 'bold',
          lineHeight: 'tight',
        },
        '6': {
          fontSize: 'lg',
          fontWeight: 'bold',
          lineHeight: 'tight',
        },
      },
    }
	},
	plugins: [
		catppuccin({
      prefix: "ctp",
    }),
		plugin(function ({ addVariant, matchUtilities, theme }) {
      addVariant('hocus', ['&:hover', '&:focus']);
      // Square utility
      matchUtilities(
        {
          square: (value) => ({
            width: value,
            height: value,
          }),
        },
        { values: theme('spacing') }
      );
    }),
	],
} satisfies Config;

module.exports = config;
