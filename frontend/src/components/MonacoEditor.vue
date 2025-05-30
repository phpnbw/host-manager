<template>
  <div ref="editorContainer" class="monaco-editor-container"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as monaco from 'monaco-editor'

interface Props {
  modelValue: string
  language?: string
  theme?: string
  readonly?: boolean
  height?: string
}

interface Emits {
  (e: 'update:modelValue', value: string): void
}

const props = withDefaults(defineProps<Props>(), {
  language: 'plaintext',
  theme: 'vs-dark',
  readonly: false,
  height: '400px'
})

const emit = defineEmits<Emits>()

const editorContainer = ref<HTMLElement>()
let editor: monaco.editor.IStandaloneCodeEditor | null = null

// 根据文件扩展名获取语言
const getLanguageFromFilename = (filename: string): string => {
  const ext = filename.toLowerCase().split('.').pop()
  const languageMap: Record<string, string> = {
    'js': 'javascript',
    'ts': 'typescript',
    'jsx': 'javascript',
    'tsx': 'typescript',
    'vue': 'html',
    'html': 'html',
    'htm': 'html',
    'css': 'css',
    'scss': 'scss',
    'sass': 'sass',
    'less': 'less',
    'json': 'json',
    'xml': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    'md': 'markdown',
    'py': 'python',
    'go': 'go',
    'java': 'java',
    'c': 'c',
    'cpp': 'cpp',
    'h': 'c',
    'hpp': 'cpp',
    'php': 'php',
    'rb': 'ruby',
    'sh': 'shell',
    'bash': 'shell',
    'zsh': 'shell',
    'sql': 'sql',
    'dockerfile': 'dockerfile',
    'conf': 'ini',
    'config': 'ini',
    'ini': 'ini',
    'env': 'shell',
    'log': 'plaintext',
    'txt': 'plaintext'
  }
  
  return languageMap[ext || ''] || 'plaintext'
}

const initEditor = () => {
  if (!editorContainer.value) return

  editor = monaco.editor.create(editorContainer.value, {
    value: props.modelValue,
    language: props.language,
    theme: props.theme,
    readOnly: props.readonly,
    automaticLayout: true,
    fontSize: 14,
    lineNumbers: 'on',
    minimap: { enabled: false },
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    folding: true,
    lineDecorationsWidth: 10,
    lineNumbersMinChars: 3
  })

  // 监听内容变化
  editor.onDidChangeModelContent(() => {
    if (editor) {
      emit('update:modelValue', editor.getValue())
    }
  })
}

const updateLanguage = (filename: string) => {
  if (editor) {
    const language = getLanguageFromFilename(filename)
    const model = editor.getModel()
    if (model) {
      monaco.editor.setModelLanguage(model, language)
    }
  }
}

// 监听modelValue变化
watch(() => props.modelValue, (newValue) => {
  if (editor && editor.getValue() !== newValue) {
    editor.setValue(newValue)
  }
})

// 监听language变化
watch(() => props.language, (newLanguage) => {
  if (editor) {
    const model = editor.getModel()
    if (model) {
      monaco.editor.setModelLanguage(model, newLanguage)
    }
  }
})

onMounted(() => {
  initEditor()
})

onUnmounted(() => {
  if (editor) {
    editor.dispose()
  }
})

// 暴露方法给父组件
defineExpose({
  updateLanguage,
  getEditor: () => editor
})
</script>

<style scoped>
.monaco-editor-container {
  width: 100%;
  height: v-bind(height);
  border: 1px solid #d9d9d9;
  border-radius: 4px;
}
</style> 