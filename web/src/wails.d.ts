/** Wails 桌面环境绑定类型声明。 */
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetAPIPort: () => Promise<number>
        }
      }
    }
  }
}

export {}
