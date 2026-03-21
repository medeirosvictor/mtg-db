/** @type {import('tailwindcss').Config} */

function withOpacity(varName) {
  return ({ opacityValue }) => {
    if (opacityValue !== undefined) {
      return `rgb(var(${varName}) / ${opacityValue})`;
    }
    return `rgb(var(${varName}))`;
  };
}

export default {
  content: [
    './index.html',
    './src/**/*.{svelte,js,ts,jsx,tsx}'
  ],
  theme: {
    extend: {
      colors: {
        'bg-primary': withOpacity('--bg-primary'),
        'bg-secondary': withOpacity('--bg-secondary'),
        'bg-surface': withOpacity('--bg-surface'),
        'bg-hover': withOpacity('--bg-hover'),
        'text-primary': withOpacity('--text-primary'),
        'text-secondary': withOpacity('--text-secondary'),
        'text-muted': withOpacity('--text-muted'),
        'accent': withOpacity('--accent'),
        'accent-hover': withOpacity('--accent-hover'),
        'green': withOpacity('--green'),
        'yellow': withOpacity('--yellow'),
        'red': withOpacity('--red'),
        'orange': withOpacity('--orange'),
        'mauve': withOpacity('--mauve'),
        'border': withOpacity('--border'),
      },
      borderRadius: {
        'DEFAULT': '3px',
        'lg': '5px',
      },
    },
  },
  plugins: [],
}
