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
		extend: {}
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
