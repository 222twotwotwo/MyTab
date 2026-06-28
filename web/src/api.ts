import type { ProjectManifest } from './types'

const devApiBase = `${window.location.protocol}//${window.location.hostname}:8080`

export const API_BASE_URL = (
  import.meta.env.VITE_API_BASE_URL ||
  (window.location.port === '5173' ? devApiBase : window.location.origin)
).replace(/\/$/, '')

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, init)
  if (!response.ok) {
    let message = `HTTP ${response.status}`
    try {
      const body = (await response.json()) as { error?: string }
      message = body.error || message
    } catch {
      // Keep the status-based message when the body is not JSON.
    }
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

export function listProjects(): Promise<ProjectManifest[]> {
  return request<ProjectManifest[]>('/api/projects')
}

export function uploadProject(name: string, files: File[]): Promise<ProjectManifest> {
  const formData = new FormData()
  formData.set('name', name)
  files.forEach((file) => {
    formData.append('files', file)
    formData.append('paths', file.webkitRelativePath || file.name)
  })

  return request<ProjectManifest>('/api/projects', {
    method: 'POST',
    body: formData
  })
}

export function previewUrl(project: ProjectManifest): string {
  return `${API_BASE_URL}/api/projects/${project.id}/preview/${encodePath(project.entryFile)}`
}

function encodePath(value: string): string {
  return value
    .split('/')
    .map((part) => encodeURIComponent(part))
    .join('/')
}
