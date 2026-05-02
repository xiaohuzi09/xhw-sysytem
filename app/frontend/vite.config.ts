import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import UnoCSS from 'unocss/vite'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    UnoCSS(),
  ],
  resolve: {
    alias: {
      '@wailsjs': path.resolve(__dirname, './wailsjs'),
    },
  },
  server: {
    port: parseInt(process.env.WAILS_VITE_PORT || '9245'),
    strictPort: true,
  },
});
