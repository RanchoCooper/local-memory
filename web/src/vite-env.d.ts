/// <reference types="vite/client" />

declare module '*.css' {
  const content: string
  export default content
}

declare module '@xyflow/react/dist/style.css' {
  const content: string
  export default content
}