# React Go Admin - Web

前端基于 **React 19 + Vite 8 + shadcn/ui + Tailwind CSS 4 + Bun**。

## 技术栈

- React 19
- React Router 7
- Vite 8
- shadcn/ui
- Tailwind CSS 4
- Axios
- Bun 1.3+

## 目录结构

```text
web/
├── public/
├── src/
│   ├── api/
│   ├── components/
│   ├── hooks/
│   ├── pages/
│   ├── router/
│   └── utils/
├── package.json
└── vite.config.js
```

## 安装与运行

安装依赖：

```bash
bun install
```

启动开发服务器：

```bash
bun run dev
```

构建生产版本：

```bash
bun run build
```

运行代码检查：

```bash
bun run lint
```

预览构建产物：

```bash
bun run preview
```

## 开发说明

- 开发环境默认运行在 `http://127.0.0.1:5173`
- Vite 代理会将 `/api` 请求转发到后端 `http://127.0.0.1:9999`
- 页面代码位于 `src/pages/`
- 复用 UI 组件位于 `src/components/`
- API 封装位于 `src/api/index.js`

## 与后端联调

先启动后端：

```bash
go run ./app
```

再启动前端：

```bash
cd web
bun run dev
```
