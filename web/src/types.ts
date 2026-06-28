export interface StoredFile {
  name: string
  size: number
  kind: 'html' | 'css' | 'javascript' | 'json' | 'text' | 'image' | 'font' | 'media'
}

export interface ProjectManifest {
  id: string
  name: string
  createdAt: string
  entryFile: string
  files: StoredFile[]
}
