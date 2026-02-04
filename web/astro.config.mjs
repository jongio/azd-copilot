import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';
import mdx from '@astrojs/mdx';

// https://astro.build/config
export default defineConfig({
  site: 'https://jongio.github.io/azd-copilot/',
  base: '/azd-copilot/',
  integrations: [
    mdx()
  ],
  vite: {
    plugins: [tailwindcss()]
  },
  output: 'static'
});
