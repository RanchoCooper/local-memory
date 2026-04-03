# LocalMemory Web UI

A modern React-based web interface for LocalMemory, featuring a 3D neural graph visualization of your memory network.

## Tech Stack

| Category | Technology | Rationale |
|----------|------------|-----------|
| Framework | React 18 + TypeScript | Type safety, component-based architecture |
| Build Tool | Vite 5 | Fast HMR, optimized builds |
| Styling | TailwindCSS | Utility-first, consistent design |
| State Management | Zustand | Lightweight, minimal boilerplate |
| HTTP Client | Axios | Interceptors, request/response handling |
| 3D Graphics | Three.js | WebGL-powered 3D visualization |
| Icons | Lucide React | Clean, consistent icon set |
| Routing | React Router v6 | SPA navigation |

## Project Structure

```
web/
├── public/
│   └── favicon.svg
├── src/
│   ├── api/                    # API client layer
│   │   ├── client.ts           # Axios instance with interceptors
│   │   ├── memories.ts         # Memory CRUD operations
│   │   └── types.ts            # TypeScript interfaces
│   │
│   ├── components/             # React components
│   │   └── MemoryGraph3D.tsx   # 3D neural graph visualization
│   │
│   ├── stores/                 # Zustand state management
│   │   └── memoryStore.ts      # Memory state and actions
│   │
│   ├── App.tsx                # Main application component
│   ├── main.tsx               # Entry point
│   └── index.css              # Global styles and CSS variables
│
├── index.html
├── package.json
├── vite.config.ts
├── tailwind.config.js
├── tsconfig.json
└── postcss.config.js
```

## Code Organization

### API Layer (`src/api/`)

The API layer provides a clean interface to the LocalMemory backend:

```typescript
// client.ts - Axios instance with base configuration
export const memoriesApi = {
  list(params)     // GET /api/v1/memories
  get(id)           // GET /api/v1/memories/:id
  create(data)      // POST /api/v1/memories
  update(id, data)  // PUT /api/v1/memories/:id
  delete(id)        // DELETE /api/v1/memories/:id
  search(query)     // POST /api/v1/query
  stats()           // GET /api/v1/stats
}
```

### Components (`src/components/`)

| Component | Description |
|-----------|-------------|
| `MemoryGraph3D` | Three.js 3D visualization of memory network with drag/zoom/hover interactions |

### Pages (Integrated in `App.tsx`)

| Page | Route | Description |
|------|-------|-------------|
| Dashboard | `/` | Stats overview + recent memories |
| Memories | `/memories` | Full list with type filtering |
| Neural Graph | `/graph` | 3D visualization of memory connections |

## Key Features

### 3D Neural Graph

The Neural Graph page uses Three.js to render memories as interconnected spheres in 3D space:

- **Drag** to rotate the view
- **Scroll** to zoom in/out
- **Hover** over nodes to see memory details
- **Color-coded** by memory type:
  - Preference: Blue `#3b82f6`
  - Fact: Green `#10b981`
  - Event: Yellow `#f59e0b`
  - Skill: Purple `#8b5cf6`
  - Goal: Orange `#f97316`
  - Relationship: Pink `#ec4899`

### Design System

The UI follows a "Neural Digital" aesthetic:

- **Dark theme** with cyan/purple glow effects
- **Glass morphism** cards with backdrop blur
- **Smooth animations** for state transitions
- **Custom fonts**: Oxanium (display), Space Mono (code)

## Getting Started

### Prerequisites

- Node.js 18+
- npm or pnpm
- LocalMemory backend running on port 8080

### Installation

```bash
cd web
npm install
```

### Development

```bash
# Start development server (proxies API to localhost:8080)
npm run dev
```

The app will be available at http://localhost:5173

### Build

```bash
# Production build
npm run build

# Preview production build
npm run preview
```

## API Integration

The frontend communicates with the LocalMemory HTTP API running on port 8080. Vite proxies `/api/*` requests to avoid CORS issues:

```typescript
// vite.config.ts
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true,
    },
  },
}
```

## Memory Data Model

```typescript
interface Memory {
  id: string
  type: 'preference' | 'fact' | 'event' | 'skill' | 'goal' | 'relationship'
  scope: 'global' | 'session' | 'agent'
  key: string
  value: string
  confidence: number
  tags?: string[]
  related_ids?: string[]
  created_at: number
  updated_at: number
  deleted_at?: number
}
```

## Styling

Global styles are defined in `src/index.css` with CSS custom properties:

```css
:root {
  --bg-deep: #030712;
  --bg-primary: #0a0f1a;
  --glow-primary: #06b6d4;
  --font-display: 'Oxanium', sans-serif;
  --font-mono: 'Space Mono', monospace;
}
```

## Browser Support

- Modern browsers with WebGL support
- Chrome/Edge recommended for best 3D performance
