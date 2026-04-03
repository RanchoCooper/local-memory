import { useState, useEffect, useCallback } from 'react'
import type { Memory, StatsResponse } from './api/types'
import { memoriesApi } from './api/client'
import { MemoryGraph3D } from './components/MemoryGraph3D'
import './index.css'

// Particles component for background effect
function Particles() {
  const particles = Array.from({ length: 20 }, (_, i) => ({
    id: i,
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    delay: `${Math.random() * 6}s`,
    size: 2 + Math.random() * 4,
  }))

  return (
    <div className="neural-bg">
      {particles.map((p) => (
        <div
          key={p.id}
          className="particle"
          style={{
            left: p.left,
            top: p.top,
            width: p.size,
            height: p.size,
            animationDelay: p.delay,
            animationDuration: `${4 + Math.random() * 4}s`,
          }}
        />
      ))}
    </div>
  )
}

// Memory type colors
const typeColors: Record<string, string> = {
  preference: '#3b82f6',
  fact: '#10b981',
  event: '#f59e0b',
  skill: '#8b5cf6',
  goal: '#f97316',
  relationship: '#ec4899',
}

// Icons
const Icons = {
  Brain: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 4.5a2.5 2.5 0 0 0-4.96-.46 2.5 2.5 0 0 0-1.98 3 2.5 2.5 0 0 0 1.32 4.24 3 3 0 0 0 .34 5.02 2.5 2.5 0 0 0 5.96.44H12a2.5 2.5 0 0 0 2.5-2.5 2.5 2.5 0 0 0-2.5-2.5Z"/>
      <path d="M12 4.5a2.5 2.5 0 0 1 4.96-.46 2.5 2.5 0 0 1 1.98 3 2.5 2.5 0 0 1-1.32 4.24 3 3 0 0 1-.34 5.02 2.5 2.5 0 0 1-5.96.44H12a2.5 2.5 0 0 1 0-5 2.5 2.5 0 0 1 2.5-2.5Z"/>
      <path d="M12 4.5V18"/>
    </svg>
  ),
  Plus: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 5v14M5 12h14"/>
    </svg>
  ),
  Search: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="11" cy="11" r="8"/>
      <path d="M21 21l-4.35-4.35"/>
    </svg>
  ),
  Trash: () => (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
    </svg>
  ),
  Network: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="5" r="3"/>
      <circle cx="5" cy="19" r="3"/>
      <circle cx="19" cy="19" r="3"/>
      <path d="M12 8v4M8.5 16.5l-2-2M15.5 16.5l2-2"/>
    </svg>
  ),
  Grid: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="3" y="3" width="7" height="7" rx="1"/>
      <rect x="14" y="3" width="7" height="7" rx="1"/>
      <rect x="3" y="14" width="7" height="7" rx="1"/>
      <rect x="14" y="14" width="7" height="7" rx="1"/>
    </svg>
  ),
  Home: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
      <polyline points="9,22 9,12 15,12 15,22"/>
    </svg>
  ),
  X: () => (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M18 6L6 18M6 6l12 12"/>
    </svg>
  ),
  Clock: () => (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10"/>
      <path d="M12 6v6l4 2"/>
    </svg>
  ),
}

// Navigation component
function Nav({ active, onNavigate }: { active: string; onNavigate: (v: string) => void }) {
  const navItems = [
    { id: 'dashboard', label: 'Memory Palace', icon: Icons.Home },
    { id: 'memories', label: 'All Memories', icon: Icons.Grid },
    { id: 'graph', label: 'Neural Graph', icon: Icons.Network },
  ]

  return (
    <nav style={{
      position: 'fixed',
      top: 0,
      left: 0,
      right: 0,
      zIndex: 100,
      background: 'rgba(3, 7, 18, 0.8)',
      backdropFilter: 'blur(12px)',
      borderBottom: '1px solid rgba(148, 163, 184, 0.1)',
    }}>
      <div style={{
        maxWidth: 1200,
        margin: '0 auto',
        padding: '0 24px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        height: 64,
      }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{
            width: 36,
            height: 36,
            borderRadius: '50%',
            background: 'linear-gradient(135deg, #06b6d4, #8b5cf6)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}>
            <Icons.Brain />
          </div>
          <span className="display-text" style={{ fontSize: 20, fontWeight: 700 }}>
            LocalMemory
          </span>
        </div>

        <div style={{ display: 'flex', gap: 4 }}>
          {navItems.map(({ id, label, icon: Icon }) => (
            <button
              key={id}
              onClick={() => onNavigate(id)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                padding: '8px 16px',
                borderRadius: 8,
                border: 'none',
                cursor: 'pointer',
                fontFamily: 'var(--font-display)',
                fontSize: 14,
                fontWeight: 500,
                transition: 'all 0.2s ease',
                background: active === id ? 'rgba(6, 182, 212, 0.15)' : 'transparent',
                color: active === id ? '#06b6d4' : '#94a3b8',
              }}
            >
              <Icon />
              {label}
            </button>
          ))}
        </div>
      </div>
    </nav>
  )
}

// Stats card component
function StatCard({ label, value, color, delay }: { label: string; value: number; color: string; delay: number }) {
  return (
    <div className="glass-card" style={{ padding: 24, position: 'relative', overflow: 'hidden' }}>
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        width: 4,
        height: '100%',
        background: color,
        borderRadius: '4px 0 0 4px',
      }} />
      <div style={{ paddingLeft: 16 }}>
        <p style={{ margin: 0, fontSize: 12, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          {label}
        </p>
        <p
          className="stat-value"
          style={{
            margin: '8px 0 0',
            fontSize: 36,
            fontWeight: 700,
            color,
            fontFamily: 'var(--font-mono)',
            animationDelay: `${delay}ms`,
          }}
        >
          {value}
        </p>
      </div>
    </div>
  )
}

// Memory card component
function MemoryCard({ memory, onDelete, onClick }: {
  memory: Memory
  onDelete: (id: string) => void
  onClick: () => void
}) {
  const color = typeColors[memory.type] || '#06b6d4'

  return (
    <div
      className="memory-node"
      onClick={onClick}
      style={{ '--node-color': color } as React.CSSProperties}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 12 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span className={`type-dot ${memory.type}`} />
          <span style={{ fontSize: 11, color: '#64748b', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
            {memory.type}
          </span>
        </div>
        <span style={{
          fontSize: 10,
          padding: '2px 8px',
          borderRadius: 10,
          background: 'rgba(6, 182, 212, 0.1)',
          color: '#06b6d4',
          border: '1px solid rgba(6, 182, 212, 0.2)',
        }}>
          {memory.scope}
        </span>
      </div>

      <h3 style={{
        margin: '0 0 8px',
        fontSize: 16,
        fontWeight: 600,
        color: '#f1f5f9',
        fontFamily: 'var(--font-display)',
      }}>
        {memory.key}
      </h3>

      <p style={{
        margin: '0 0 12px',
        fontSize: 14,
        color: '#94a3b8',
        lineHeight: 1.5,
        display: '-webkit-box',
        WebkitLineClamp: 2,
        WebkitBoxOrient: 'vertical',
        overflow: 'hidden',
      }}>
        {memory.value}
      </p>

      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: '#64748b', fontSize: 12 }}>
          <Icons.Clock />
          {new Date(memory.updated_at * 1000).toLocaleDateString('en-US', {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
          })}
        </div>

        {(memory.tags?.length ?? 0) > 0 && (
          <div style={{ display: 'flex', gap: 4 }}>
            {memory.tags.slice(0, 2).map((tag) => (
              <span key={tag} className="tag">{tag}</span>
            ))}
          </div>
        )}
      </div>

      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation()
            onDelete(memory.id)
          }}
          style={{
            position: 'absolute',
            bottom: 12,
            right: 12,
            width: 28,
            height: 28,
            borderRadius: 6,
            border: 'none',
            background: 'rgba(239, 68, 68, 0.1)',
            color: '#ef4444',
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            opacity: 0,
            transition: 'opacity 0.2s ease',
          }}
          className="delete-btn"
        >
          <Icons.Trash />
        </button>
      )}
    </div>
  )
}

// Memory form modal
function MemoryFormModal({ isOpen, onClose, onSubmit }: {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: { type: string; scope: string; key: string; value: string; tags: string }) => void
}) {
  const [form, setForm] = useState({
    type: 'fact',
    scope: 'global',
    key: '',
    value: '',
    tags: '',
  })

  if (!isOpen) return null

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 24 }}>
          <h2 style={{ margin: 0, fontSize: 20, fontWeight: 600 }}>Store New Memory</h2>
          <button
            onClick={onClose}
            style={{
              width: 32,
              height: 32,
              borderRadius: 8,
              border: '1px solid rgba(6, 182, 212, 0.3)',
              background: 'rgba(6, 182, 212, 0.1)',
              color: '#06b6d4',
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              transition: 'all 0.2s ease',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.background = 'rgba(6, 182, 212, 0.2)'
              e.currentTarget.style.borderColor = '#06b6d4'
              e.currentTarget.style.boxShadow = '0 0 12px rgba(6, 182, 212, 0.3)'
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.background = 'rgba(6, 182, 212, 0.1)'
              e.currentTarget.style.borderColor = 'rgba(6, 182, 212, 0.3)'
              e.currentTarget.style.boxShadow = 'none'
            }}
          >
            <Icons.X />
          </button>
        </div>

        <form onSubmit={(e) => {
          e.preventDefault()
          onSubmit(form)
        }} style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div>
              <label style={{ display: 'block', fontSize: 12, color: '#94a3b8', marginBottom: 6, textTransform: 'uppercase' }}>
                Type
              </label>
              <select
                value={form.type}
                onChange={(e) => setForm({ ...form, type: e.target.value })}
                className="input"
              >
                {Object.keys(typeColors).map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
            <div>
              <label style={{ display: 'block', fontSize: 12, color: '#94a3b8', marginBottom: 6, textTransform: 'uppercase' }}>
                Scope
              </label>
              <select
                value={form.scope}
                onChange={(e) => setForm({ ...form, scope: e.target.value })}
                className="input"
              >
                <option value="global">global</option>
                <option value="session">session</option>
                <option value="agent">agent</option>
              </select>
            </div>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: 12, color: '#94a3b8', marginBottom: 6, textTransform: 'uppercase' }}>
              Key
            </label>
            <input
              type="text"
              value={form.key}
              onChange={(e) => setForm({ ...form, key: e.target.value })}
              className="input"
              placeholder="e.g., language_preference"
              required
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: 12, color: '#94a3b8', marginBottom: 6, textTransform: 'uppercase' }}>
              Value
            </label>
            <textarea
              value={form.value}
              onChange={(e) => setForm({ ...form, value: e.target.value })}
              className="input"
              placeholder="e.g., User prefers Go language for backend development"
              rows={4}
              required
              style={{ resize: 'vertical' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: 12, color: '#94a3b8', marginBottom: 6, textTransform: 'uppercase' }}>
              Tags (comma separated)
            </label>
            <input
              type="text"
              value={form.tags}
              onChange={(e) => setForm({ ...form, tags: e.target.value })}
              className="input"
              placeholder="e.g., golang, backend, programming"
            />
          </div>

          <div style={{ display: 'flex', gap: 12, marginTop: 8 }}>
            <button type="button" onClick={onClose} className="btn btn-secondary" style={{ flex: 1 }}>
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" style={{ flex: 1 }}>
              Store Memory
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

// Dashboard component
function Dashboard({ stats, memories, onDelete, onAddClick }: {
  stats: StatsResponse | null
  memories: Memory[]
  onDelete: (id: string) => void
  onAddClick: () => void
}) {
  const recentMemories = memories.slice(0, 5)

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 32 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h1 className="display-text" style={{ margin: 0, fontSize: 32 }}>Memory Palace</h1>
          <p style={{ margin: '8px 0 0', color: '#64748b' }}>Your neural memory network</p>
        </div>
        <button onClick={onAddClick} className="btn btn-primary">
          <Icons.Plus />
          New Memory
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 16 }}>
        <StatCard label="Total" value={stats?.total ?? 0} color="#06b6d4" delay={0} />
        <StatCard label="Global" value={stats?.by_scope?.global ?? 0} color="#10b981" delay={100} />
        <StatCard label="Session" value={stats?.by_scope?.session ?? 0} color="#f59e0b" delay={200} />
        <StatCard label="Agent" value={stats?.by_scope?.agent ?? 0} color="#8b5cf6" delay={300} />
      </div>

      {recentMemories.length > 0 ? (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <h2 style={{ margin: 0, fontSize: 18, fontWeight: 600 }}>Recent Memories</h2>
          {recentMemories.map((memory) => (
            <MemoryCard key={memory.id} memory={memory} onDelete={onDelete} onClick={() => {}} />
          ))}
        </div>
      ) : (
        <div style={{
          padding: 48,
          textAlign: 'center',
          background: 'var(--bg-card)',
          borderRadius: 16,
          border: '1px solid var(--border-subtle)',
        }}>
          <p style={{ margin: 0, color: '#64748b', fontSize: 16 }}>No memories yet. Start storing your first memory!</p>
        </div>
      )}
    </div>
  )
}

// Memories list component
function MemoriesList({ memories, onDelete, onAddClick }: {
  memories: Memory[]
  onDelete: (id: string) => void
  onAddClick: () => void
}) {
  const [filter, setFilter] = useState('all')

  const filteredMemories = filter === 'all'
    ? memories
    : memories.filter((m) => m.type === filter)

  const types = ['all', ...Object.keys(typeColors)]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div>
          <h1 className="display-text" style={{ margin: 0, fontSize: 32 }}>All Memories</h1>
          <p style={{ margin: '8px 0 0', color: '#64748b' }}>{memories.length} memories stored</p>
        </div>
        <button onClick={onAddClick} className="btn btn-primary">
          <Icons.Plus />
          New Memory
        </button>
      </div>

      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {types.map((type) => (
          <button
            key={type}
            onClick={() => setFilter(type)}
            style={{
              padding: '6px 14px',
              borderRadius: 20,
              border: 'none',
              cursor: 'pointer',
              fontFamily: 'var(--font-mono)',
              fontSize: 12,
              transition: 'all 0.2s ease',
              background: filter === type ? 'var(--glow-primary)' : 'var(--bg-elevated)',
              color: filter === type ? '#030712' : '#94a3b8',
            }}
          >
            {type}
          </button>
        ))}
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(340px, 1fr))', gap: 16 }}>
        {filteredMemories.map((memory) => (
          <MemoryCard key={memory.id} memory={memory} onDelete={onDelete} onClick={() => {}} />
        ))}
      </div>

      {filteredMemories.length === 0 && (
        <div style={{
          padding: 48,
          textAlign: 'center',
          background: 'var(--bg-card)',
          borderRadius: 16,
          border: '1px solid var(--border-subtle)',
        }}>
          <p style={{ margin: 0, color: '#64748b', fontSize: 16 }}>No memories found</p>
        </div>
      )}
    </div>
  )
}

// Graph page
function GraphPage({ memories }: { memories: Memory[] }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
      <h1 className="display-text" style={{ margin: 0, fontSize: 32 }}>Neural Graph</h1>
      <p style={{ margin: 0, color: '#64748b' }}>
        Visualize connections between your memories
      </p>
      <MemoryGraph3D memories={memories} />
    </div>
  )
}

// Main App
export default function App() {
  const [page, setPage] = useState('dashboard')
  const [memories, setMemories] = useState<Memory[]>([])
  const [stats, setStats] = useState<StatsResponse | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [loading, setLoading] = useState(true)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [memResp, statsResp] = await Promise.all([
        memoriesApi.list({ limit: 100 }),
        memoriesApi.stats(),
      ])
      if (memResp.success && memResp.data) {
        setMemories(memResp.data.memories)
      }
      if (statsResp.success && statsResp.data) {
        setStats(statsResp.data)
      }
    } catch (err) {
      console.error('Failed to fetch data:', err)
    }
    setLoading(false)
  }, [])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const handleDelete = async (id: string) => {
    try {
      await memoriesApi.delete(id)
      setMemories((prev) => prev.filter((m) => m.id !== id))
      fetchData()
    } catch (err) {
      console.error('Failed to delete:', err)
    }
  }

  const handleSubmit = async (data: { type: string; scope: string; key: string; value: string; tags: string }) => {
    try {
      await memoriesApi.create({
        type: data.type as Memory['type'],
        scope: data.scope as Memory['scope'],
        key: data.key,
        value: data.value,
        tags: data.tags ? data.tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
      })
      setShowForm(false)
      fetchData()
    } catch (err) {
      console.error('Failed to create:', err)
    }
  }

  return (
    <>
      <Particles />
      <Nav active={page} onNavigate={setPage} />

      <main style={{
        paddingTop: 96,
        paddingBottom: 48,
        minHeight: '100vh',
        position: 'relative',
        zIndex: 1,
      }}>
        <div style={{ maxWidth: 1200, margin: '0 auto', padding: '0 24px' }}>
          {loading ? (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '60vh' }}>
              <div className="pulse-glow" style={{
                width: 48,
                height: 48,
                borderRadius: '50%',
                background: 'linear-gradient(135deg, #06b6d4, #8b5cf6)',
              }} />
            </div>
          ) : (
            <>
              {page === 'dashboard' && (
                <Dashboard
                  stats={stats}
                  memories={memories}
                  onDelete={handleDelete}
                  onAddClick={() => setShowForm(true)}
                />
              )}
              {page === 'memories' && (
                <MemoriesList
                  memories={memories}
                  onDelete={handleDelete}
                  onAddClick={() => setShowForm(true)}
                />
              )}
              {page === 'graph' && <GraphPage memories={memories} />}
            </>
          )}
        </div>
      </main>

      <MemoryFormModal
        isOpen={showForm}
        onClose={() => setShowForm(false)}
        onSubmit={handleSubmit}
      />

      <style>{`
        .memory-node:hover .delete-btn { opacity: 1 !important; }
      `}</style>
    </>
  )
}