<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import {
  AlertCircle,
  CheckCircle2,
  FileText,
  FolderUp,
  PanelTop,
  RefreshCw,
  Settings,
  Store,
  UploadCloud
} from 'lucide-vue-next'
import { listProjects, previewUrl, uploadProject } from './api'
import { marketStyles, type MarketStyle } from './marketStyles'
import type { ProjectManifest } from './types'

const activeProjectKey = 'mytab.activeProjectId'
const acceptedExtensions = [
  '.html',
  '.htm',
  '.css',
  '.js',
  '.mjs',
  '.json',
  '.txt',
  '.png',
  '.jpg',
  '.jpeg',
  '.gif',
  '.webp',
  '.svg',
  '.ico',
  '.woff',
  '.woff2',
  '.ttf',
  '.otf',
  '.mp3',
  '.mp4',
  '.webm'
]

const projects = ref<ProjectManifest[]>([])
const selectedId = ref('')
const activeId = ref(localStorage.getItem(activeProjectKey) || '')
const selectedFiles = ref<File[]>([])
const projectName = ref('')
const activeManagerTab = ref<'uploads' | 'market'>('uploads')
const selectedStyleId = ref(marketStyles[0]?.id || '')
const loading = ref(false)
const uploading = ref(false)
const applyingStyle = ref(false)
const managing = ref(!activeId.value)
const error = ref('')
const notice = ref('')
const frameKey = ref(0)

const selectedProject = computed(() => projects.value.find((project) => project.id === selectedId.value))
const activeProject = computed(() => projects.value.find((project) => project.id === activeId.value))
const selectedStyle = computed(() => marketStyles.find((style) => style.id === selectedStyleId.value) || marketStyles[0])
const visibleProject = computed(() => (managing.value ? selectedProject.value : activeProject.value))
const selectedPreviewUrl = computed(() => (selectedProject.value ? previewUrl(selectedProject.value) : ''))
const visiblePreviewUrl = computed(() => (visibleProject.value ? previewUrl(visibleProject.value) : ''))
const selectedStylePreviewUrl = computed(() => (selectedStyle.value ? stylePreviewUrl(selectedStyle.value) : ''))
const canUpload = computed(() => selectedFiles.value.length > 0 && !uploading.value)
const hasAppliedProject = computed(() => Boolean(activeProject.value))

onMounted(() => {
  void refreshProjects()
})

async function refreshProjects() {
  loading.value = true
  error.value = ''
  try {
    projects.value = await listProjects()
    if (activeId.value && !projects.value.some((project) => project.id === activeId.value)) {
      activeId.value = ''
      localStorage.removeItem(activeProjectKey)
      managing.value = true
    }
    if (!selectedId.value && projects.value.length > 0) {
      selectedId.value = activeId.value || projects.value[0].id
    }
  } catch (err) {
    error.value = toMessage(err)
    managing.value = true
  } finally {
    loading.value = false
  }
}

async function submitUpload() {
  if (!canUpload.value) {
    return
  }

  error.value = ''
  notice.value = ''

  const validationError = validateFiles(selectedFiles.value)
  if (validationError) {
    error.value = validationError
    return
  }

  uploading.value = true
  try {
    const created = await uploadProject(projectName.value, selectedFiles.value)
    projects.value = [created, ...projects.value.filter((project) => project.id !== created.id)]
    selectedId.value = created.id
    selectedFiles.value = []
    projectName.value = ''
    applyProject(created.id)
    notice.value = '已上传并应用到当前页面'
  } catch (err) {
    error.value = toMessage(err)
  } finally {
    uploading.value = false
  }
}

async function applySelectedStyle() {
  if (!selectedStyle.value || applyingStyle.value) {
    return
  }

  error.value = ''
  notice.value = ''
  applyingStyle.value = true
  try {
    const file = new File([selectedStyle.value.html], 'index.html', { type: 'text/html' })
    const created = await uploadProject(`市场样式 - ${selectedStyle.value.name}`, [file])
    projects.value = [created, ...projects.value.filter((project) => project.id !== created.id)]
    selectedId.value = created.id
    applyProject(created.id)
    notice.value = '已从样式市场应用到当前页面'
  } catch (err) {
    error.value = toMessage(err)
  } finally {
    applyingStyle.value = false
  }
}

function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  selectedFiles.value = Array.from(input.files || [])
  notice.value = ''
  error.value = ''
}

function chooseProject(id: string) {
  selectedId.value = id
  frameKey.value += 1
}

function chooseStyle(id: string) {
  selectedStyleId.value = id
  frameKey.value += 1
}

function applySelectedProject() {
  if (selectedProject.value) {
    applyProject(selectedProject.value.id)
    notice.value = '已应用到当前页面'
  }
}

function openTab(tab: 'uploads' | 'market') {
  activeManagerTab.value = tab
}

function applyProject(id: string) {
  activeId.value = id
  localStorage.setItem(activeProjectKey, id)
  managing.value = false
  frameKey.value += 1
}

function openManager() {
  managing.value = true
  if (activeId.value) {
    selectedId.value = activeId.value
  }
}

function reloadCurrentPage() {
  frameKey.value += 1
}

function validateFiles(files: File[]): string {
  if (!files.some((file) => ['.html', '.htm'].includes(extensionOf(file.name)))) {
    return '请至少选择一个 HTML 文件'
  }

  const invalid = files.find((file) => !acceptedExtensions.includes(extensionOf(file.name)))
  if (invalid) {
    return `${displayFileName(invalid)} 不是支持的静态资源类型`
  }

  const names = new Set<string>()
  for (const file of files) {
    const key = displayFileName(file).toLowerCase()
    if (names.has(key)) {
      return `${displayFileName(file)} 重复`
    }
    names.add(key)
  }

  return ''
}

function extensionOf(fileName: string): string {
  const index = fileName.lastIndexOf('.')
  return index >= 0 ? fileName.slice(index).toLowerCase() : ''
}

function displayFileName(file: File): string {
  return file.webkitRelativePath || file.name
}

function stylePreviewUrl(style: MarketStyle): string {
  return `data:text/html;charset=utf-8,${encodeURIComponent(style.html)}`
}

function toMessage(err: unknown): string {
  return err instanceof Error ? err.message : '请求失败'
}

function formatSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

function formatDate(value: string): string {
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(new Date(value))
}
</script>

<template>
  <main class="tab-page" :class="{ managing }">
    <iframe
      v-if="hasAppliedProject && !managing"
      :key="frameKey"
      class="active-frame"
      :src="visiblePreviewUrl"
      sandbox="allow-scripts"
      referrerpolicy="no-referrer"
      title="Current tab page"
    ></iframe>

    <div v-if="hasAppliedProject && !managing" class="floating-actions">
      <button class="icon-button glass" type="button" title="重新加载" @click="reloadCurrentPage">
        <RefreshCw :size="18" />
      </button>
      <button class="icon-button glass" type="button" title="管理标签页" @click="openManager">
        <Settings :size="18" />
      </button>
    </div>

    <section v-else class="manager-shell">
      <header class="topbar">
        <div>
          <p class="eyebrow">MyTab</p>
          <h1>个性化浏览器标签页</h1>
        </div>
        <button class="icon-button" type="button" :disabled="loading" title="刷新" @click="refreshProjects">
          <RefreshCw :size="18" :class="{ spinning: loading }" />
        </button>
      </header>

      <nav class="manager-tabs" aria-label="管理页面">
        <button
          class="tab-button"
          :class="{ active: activeManagerTab === 'uploads' }"
          type="button"
          @click="openTab('uploads')"
        >
          <PanelTop :size="17" />
          <span>我的网页</span>
        </button>
        <button
          class="tab-button"
          :class="{ active: activeManagerTab === 'market' }"
          type="button"
          @click="openTab('market')"
        >
          <Store :size="17" />
          <span>样式市场</span>
        </button>
      </nav>

      <section v-if="activeManagerTab === 'uploads'" class="workspace">
        <aside class="side-panel">
          <form class="upload-form" @submit.prevent="submitUpload">
            <div class="section-title">
              <UploadCloud :size="18" />
              <span>上传静态网页</span>
            </div>

            <label class="field">
              <span>名称</span>
              <input v-model="projectName" type="text" maxlength="80" placeholder="我的标签页" />
            </label>

            <label class="file-picker">
              <input
                type="file"
                multiple
                webkitdirectory
                accept=".html,.htm,.css,.js,.mjs,.json,.txt,.png,.jpg,.jpeg,.gif,.webp,.svg,.ico,.woff,.woff2,.ttf,.otf,.mp3,.mp4,.webm"
                @change="onFileChange"
              />
              <FolderUp :size="22" />
              <span>{{ selectedFiles.length ? `${selectedFiles.length} 个资源` : '选择网页文件夹' }}</span>
            </label>

            <div v-if="selectedFiles.length" class="picked-files">
              <span v-for="file in selectedFiles" :key="displayFileName(file)">{{ displayFileName(file) }}</span>
            </div>

            <button class="primary-button" type="submit" :disabled="!canUpload">
              <CheckCircle2 :size="17" />
              <span>{{ uploading ? '应用中' : '上传并应用' }}</span>
            </button>
          </form>

          <div class="project-list">
            <div class="section-title">
              <PanelTop :size="18" />
              <span>已保存标签页</span>
            </div>

            <button
              v-for="project in projects"
              :key="project.id"
              class="project-item"
              :class="{ active: project.id === selectedId }"
              type="button"
              @click="chooseProject(project.id)"
            >
              <FileText :size="18" />
              <span class="project-text">
                <strong>{{ project.name }}</strong>
                <small>{{ project.entryFile }} · {{ formatDate(project.createdAt) }}</small>
              </span>
            </button>

            <p v-if="!projects.length && !loading" class="empty">暂无标签页</p>
          </div>
        </aside>

        <section class="preview-panel">
          <div class="preview-toolbar">
            <div class="section-title">
              <PanelTop :size="18" />
              <span>当前页面</span>
            </div>
            <div class="toolbar-actions">
              <button
                class="small-button"
                type="button"
                :disabled="!selectedProject"
                title="应用到当前页面"
                @click="applySelectedProject"
              >
                <CheckCircle2 :size="16" />
                <span>应用</span>
              </button>
              <button class="small-button" type="button" :disabled="!selectedProject" title="重新加载" @click="reloadCurrentPage">
                <RefreshCw :size="16" />
                <span>重载</span>
              </button>
            </div>
          </div>

          <div v-if="error" class="message error">
            <AlertCircle :size="17" />
            <span>{{ error }}</span>
          </div>
          <div v-if="notice" class="message success">
            <CheckCircle2 :size="17" />
            <span>{{ notice }}</span>
          </div>

          <div v-if="selectedProject" class="manifest-row">
            <span v-for="file in selectedProject.files" :key="file.name" class="file-chip">
              {{ file.name }} <small>{{ formatSize(file.size) }}</small>
            </span>
          </div>

          <iframe
            v-if="selectedProject"
            :key="frameKey"
            class="sandbox-frame"
            :src="selectedPreviewUrl"
            sandbox="allow-scripts"
            referrerpolicy="no-referrer"
            title="Tab page preview"
          ></iframe>

          <div v-else class="preview-empty">
            <PanelTop :size="32" />
            <span>上传或选择一个静态网页</span>
          </div>
        </section>
      </section>

      <section v-else class="market-workspace">
        <aside class="market-list">
          <div class="section-title">
            <Store :size="18" />
            <span>样式市场</span>
          </div>

          <button
            v-for="style in marketStyles"
            :key="style.id"
            class="style-card"
            :class="{ active: style.id === selectedStyleId }"
            type="button"
            @click="chooseStyle(style.id)"
          >
            <span class="style-swatches" aria-hidden="true">
              <span v-for="color in style.swatches" :key="color" :style="{ background: color }"></span>
            </span>
            <span class="style-text">
              <strong>{{ style.name }}</strong>
              <small>{{ style.description }}</small>
            </span>
            <span class="style-tags">
              <span v-for="tag in style.tags" :key="tag">{{ tag }}</span>
            </span>
          </button>
        </aside>

        <section class="preview-panel">
          <div class="preview-toolbar">
            <div class="section-title">
              <PanelTop :size="18" />
              <span>{{ selectedStyle?.name || '样式预览' }}</span>
            </div>
            <button
              class="small-button"
              type="button"
              :disabled="!selectedStyle || applyingStyle"
              title="应用样式"
              @click="applySelectedStyle"
            >
              <CheckCircle2 :size="16" />
              <span>{{ applyingStyle ? '应用中' : '应用样式' }}</span>
            </button>
          </div>

          <div v-if="error" class="message error">
            <AlertCircle :size="17" />
            <span>{{ error }}</span>
          </div>
          <div v-if="notice" class="message success">
            <CheckCircle2 :size="17" />
            <span>{{ notice }}</span>
          </div>

          <iframe
            v-if="selectedStyle"
            :key="selectedStyle.id"
            class="sandbox-frame market-frame"
            :src="selectedStylePreviewUrl"
            sandbox="allow-scripts"
            referrerpolicy="no-referrer"
            title="Style market preview"
          ></iframe>
        </section>
      </section>
    </section>
  </main>
</template>
