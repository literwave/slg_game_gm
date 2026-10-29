<script setup lang="ts">
import { useRoute } from 'vue-router'

defineProps<{ title?: string; desc?: string }>()
const route = useRoute()
</script>

<template>
  <div class="page-header">
    <div>
      <h1>{{ title ?? route.meta.title }}</h1>
      <p v-if="desc ?? route.meta.desc">{{ desc ?? route.meta.desc }}</p>
    </div>
    <div class="page-actions"><slot name="actions" /></div>
  </div>
</template>
