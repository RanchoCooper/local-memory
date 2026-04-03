import { useEffect, useRef, useState, useCallback } from 'react'
import * as THREE from 'three'
import type { Memory } from '../api/types'

const typeColors: Record<string, string> = {
  preference: '#3b82f6',
  fact: '#10b981',
  event: '#f59e0b',
  skill: '#8b5cf6',
  goal: '#f97316',
  relationship: '#ec4899',
}

interface GraphNode {
  id: string
  mesh: THREE.Group
  memory: Memory
}

interface Props {
  memories: Memory[]
}

export function MemoryGraph3D({ memories }: Props) {
  const containerRef = useRef<HTMLDivElement>(null)
  const sceneRef = useRef<THREE.Scene | null>(null)
  const cameraRef = useRef<THREE.PerspectiveCamera | null>(null)
  const rendererRef = useRef<THREE.WebGLRenderer | null>(null)
  const nodesRef = useRef<GraphNode[]>([])
  const linesRef = useRef<THREE.Line[]>([])
  const raycasterRef = useRef<THREE.Raycaster>(new THREE.Raycaster())
  const mouseRef = useRef<THREE.Vector2>(new THREE.Vector2())
  const isDraggingRef = useRef(false)
  const previousMouseRef = useRef({ x: 0, y: 0 })
  const sphericalRef = useRef({ theta: 0, phi: Math.PI / 2, radius: 500 })
  const animationRef = useRef<number>(0)

  const [hoveredMemory, setHoveredMemory] = useState<Memory | null>(null)
  const [tooltipPos, setTooltipPos] = useState({ x: 0, y: 0 })

  const initThree = useCallback(() => {
    if (!containerRef.current) return

    const container = containerRef.current
    const width = container.offsetWidth
    const height = container.offsetHeight

    // Scene
    const scene = new THREE.Scene()
    scene.background = new THREE.Color(0x030712)
    sceneRef.current = scene

    // Camera
    const camera = new THREE.PerspectiveCamera(60, width / height, 1, 2000)
    camera.position.set(0, 200, 400)
    camera.lookAt(0, 0, 0)
    cameraRef.current = camera

    // Renderer
    const renderer = new THREE.WebGLRenderer({ antialias: true })
    renderer.setSize(width, height)
    renderer.setPixelRatio(window.devicePixelRatio)
    container.appendChild(renderer.domElement)
    rendererRef.current = renderer

    // Lights
    const ambientLight = new THREE.AmbientLight(0x404040, 2)
    scene.add(ambientLight)

    const pointLight = new THREE.PointLight(0x06b6d4, 1, 1000)
    pointLight.position.set(200, 200, 200)
    scene.add(pointLight)

    // Grid helper (subtle)
    const gridHelper = new THREE.GridHelper(600, 30, 0x1a2234, 0x0f172a)
    gridHelper.position.y = -100
    scene.add(gridHelper)

    // Handle resize
    const handleResize = () => {
      if (!container || !camera || !renderer) return
      const w = container.offsetWidth
      const h = container.offsetHeight
      camera.aspect = w / h
      camera.updateProjectionMatrix()
      renderer.setSize(w, h)
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      renderer.dispose()
      if (container.contains(renderer.domElement)) {
        container.removeChild(renderer.domElement)
      }
    }
  }, [])

  const createNodes = useCallback(() => {
    if (!sceneRef.current) return

    const scene = sceneRef.current

    // Clear old nodes
    nodesRef.current.forEach(({ mesh }) => scene.remove(mesh))
    nodesRef.current = []

    // Clear old lines
    linesRef.current.forEach((line) => scene.remove(line))
    linesRef.current = []

    memories.forEach((memory, i) => {
      const color = new THREE.Color(typeColors[memory.type] || '#06b6d4')

      // Node group
      const group = new THREE.Group()

      // Outer glow sphere
      const glowGeo = new THREE.SphereGeometry(28, 32, 32)
      const glowMat = new THREE.MeshBasicMaterial({
        color: color,
        transparent: true,
        opacity: 0.15,
      })
      const glowMesh = new THREE.Mesh(glowGeo, glowMat)
      group.add(glowMesh)

      // Main sphere
      const sphereGeo = new THREE.SphereGeometry(20, 32, 32)
      const sphereMat = new THREE.MeshStandardMaterial({
        color: color,
        emissive: color,
        emissiveIntensity: 0.3,
        metalness: 0.5,
        roughness: 0.3,
      })
      const sphereMesh = new THREE.Mesh(sphereGeo, sphereMat)
      group.add(sphereMesh)

      // Ring
      const ringGeo = new THREE.TorusGeometry(26, 2, 8, 32)
      const ringMat = new THREE.MeshBasicMaterial({ color: color, transparent: true, opacity: 0.5 })
      const ringMesh = new THREE.Mesh(ringGeo, ringMat)
      ringMesh.rotation.x = Math.PI / 2
      group.add(ringMesh)

      // Position in 3D space (spiral distribution)
      const phi = Math.acos(-1 + (2 * i) / memories.length)
      const theta = Math.sqrt(memories.length * Math.PI) * phi
      const radius = 150 + Math.random() * 50

      group.position.set(
        radius * Math.cos(theta) * Math.sin(phi),
        radius * Math.sin(phi) * 0.5,
        radius * Math.cos(theta) * Math.cos(phi)
      )

      // Random initial rotation
      group.rotation.set(Math.random() * Math.PI, Math.random() * Math.PI, Math.random() * Math.PI)

      // Store data
      group.userData = { memory, id: memory.id }
      scene.add(group)
      nodesRef.current.push({ id: memory.id, mesh: group, memory })
    })

    // Create connections
    const connections: [string, string][] = []
    memories.forEach((memory, i) => {
      memories.forEach((other, j) => {
        if (i >= j) return
        if (memory.type === other.type) {
          connections.push([memory.id, other.id])
        }
      })
    })

    // Draw lines
    connections.forEach(([id1, id2]) => {
      const n1 = nodesRef.current.find((n) => n.id === id1)
      const n2 = nodesRef.current.find((n) => n.id === id2)
      if (!n1 || !n2) return

      const points = [n1.mesh.position.clone(), n2.mesh.position.clone()]
      const lineGeo = new THREE.BufferGeometry().setFromPoints(points)
      const lineMat = new THREE.LineBasicMaterial({
        color: 0x06b6d4,
        transparent: true,
        opacity: 0.3,
      })
      const line = new THREE.Line(lineGeo, lineMat)
      scene.add(line)
      linesRef.current.push(line)
    })
  }, [memories])

  const animate = useCallback(() => {
    animationRef.current = requestAnimationFrame(animate)

    const camera = cameraRef.current
    const renderer = rendererRef.current
    const scene = sceneRef.current

    if (!camera || !renderer || !scene) return

    // Update camera position from spherical coordinates
    const { theta, phi, radius } = sphericalRef.current
    camera.position.x = radius * Math.sin(phi) * Math.cos(theta)
    camera.position.y = radius * Math.cos(phi)
    camera.position.z = radius * Math.sin(phi) * Math.sin(theta)
    camera.lookAt(0, 0, 0)

    // Animate nodes (slow rotation)
    nodesRef.current.forEach(({ mesh }) => {
      mesh.rotation.y += 0.002
      mesh.rotation.x += 0.001
    })

    renderer.render(scene, camera)
  }, [])

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    isDraggingRef.current = true
    previousMouseRef.current = { x: e.clientX, y: e.clientY }
  }, [])

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!containerRef.current) return

    const rect = containerRef.current.getBoundingClientRect()
    mouseRef.current.x = ((e.clientX - rect.left) / rect.width) * 2 - 1
    mouseRef.current.y = -((e.clientY - rect.top) / rect.height) * 2 + 1

    if (isDraggingRef.current) {
      const deltaX = e.clientX - previousMouseRef.current.x
      const deltaY = e.clientY - previousMouseRef.current.y

      sphericalRef.current.theta -= deltaX * 0.005
      sphericalRef.current.phi = Math.max(0.1, Math.min(Math.PI - 0.1, sphericalRef.current.phi + deltaY * 0.005))

      previousMouseRef.current = { x: e.clientX, y: e.clientY }
    } else {
      // Raycast for hover detection
      raycasterRef.current.setFromCamera(mouseRef.current, cameraRef.current!)
      const intersects = raycasterRef.current.intersectObjects(
        nodesRef.current.map((n) => n.mesh.children[1] as THREE.Object3D),
        false
      )

      if (intersects.length > 0) {
        const intersectedMesh = intersects[0].object.parent
        if (intersectedMesh?.userData.memory) {
          setHoveredMemory(intersectedMesh.userData.memory)
          setTooltipPos({ x: e.clientX, y: e.clientY })
          containerRef.current.style.cursor = 'pointer'
        }
      } else {
        setHoveredMemory(null)
        containerRef.current.style.cursor = 'grab'
      }
    }
  }, [])

  const handleMouseUp = useCallback(() => {
    isDraggingRef.current = false
  }, [])

  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault()
    sphericalRef.current.radius = Math.max(100, Math.min(1000, sphericalRef.current.radius + e.deltaY * 0.5))
  }, [])

  useEffect(() => {
    const cleanup = initThree()
    return () => cleanup?.()
  }, [initThree])

  useEffect(() => {
    createNodes()
  }, [createNodes])

  useEffect(() => {
    animate()
    return () => cancelAnimationFrame(animationRef.current)
  }, [animate])

  return (
    <div style={{ position: 'relative', width: '100%', height: 500 }}>
      <div
        ref={containerRef}
        style={{ width: '100%', height: '100%', cursor: 'grab' }}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onWheel={handleWheel}
      />

      {hoveredMemory && (
        <div
          style={{
            position: 'fixed',
            left: tooltipPos.x + 15,
            top: tooltipPos.y - 10,
            background: 'rgba(17, 24, 39, 0.95)',
            border: '1px solid rgba(6, 182, 212, 0.4)',
            borderRadius: 12,
            padding: 16,
            minWidth: 250,
            maxWidth: 320,
            backdropFilter: 'blur(12px)',
            boxShadow: '0 8px 32px rgba(6, 182, 212, 0.2)',
            zIndex: 1000,
            pointerEvents: 'none',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 8 }}>
            <span
              className={`type-dot ${hoveredMemory.type}`}
              style={{ width: 10, height: 10, borderRadius: '50%', display: 'inline-block' }}
            />
            <span style={{ fontSize: 11, color: '#64748b', textTransform: 'uppercase' }}>
              {hoveredMemory.type}
            </span>
            <span
              style={{
                fontSize: 10,
                padding: '2px 6px',
                borderRadius: 6,
                background: 'rgba(6, 182, 212, 0.1)',
                color: '#06b6d4',
              }}
            >
              {hoveredMemory.scope}
            </span>
          </div>
          <h3 style={{ margin: '0 0 8px', fontSize: 16, fontWeight: 600, color: '#f1f5f9' }}>
            {hoveredMemory.key}
          </h3>
          <p
            style={{
              margin: 0,
              fontSize: 14,
              color: '#94a3b8',
              lineHeight: 1.5,
              maxHeight: 60,
              overflow: 'hidden',
            }}
          >
            {hoveredMemory.value}
          </p>
          {hoveredMemory.tags && hoveredMemory.tags.length > 0 && (
            <div style={{ display: 'flex', gap: 4, marginTop: 12, flexWrap: 'wrap' }}>
              {hoveredMemory.tags.slice(0, 4).map((tag) => (
                <span
                  key={tag}
                  style={{
                    fontSize: 10,
                    padding: '2px 8px',
                    borderRadius: 10,
                    background: 'rgba(6, 182, 212, 0.1)',
                    color: '#06b6d4',
                  }}
                >
                  {tag}
                </span>
              ))}
            </div>
          )}
        </div>
      )}

      <div
        style={{
          position: 'absolute',
          bottom: 16,
          left: 16,
          fontSize: 11,
          color: '#64748b',
          fontFamily: 'var(--font-mono)',
        }}
      >
        Drag to rotate | Scroll to zoom | Hover for details
      </div>
    </div>
  )
}