/// <reference types="vite/client" />

export {};

declare module "vue-router" {
  interface RouteMeta {
    requiresAdmin?: boolean;
  }
}

// UnoCSS 类型声明
declare module 'virtual:uno.css' {
  const css: string
  export default css
}

interface ImportMetaEnv {
  readonly VITE_APP_ENV: string
  readonly VITE_API_BASE_URL: string
  readonly VITE_DEBUG: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
