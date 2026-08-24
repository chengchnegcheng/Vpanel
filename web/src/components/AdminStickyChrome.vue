<template>
  <div class="admin-sticky-chrome">
    <slot />
  </div>
</template>

<script setup>
// Groups title / stats / filters.
// List pages: only toolbar/filter sticks after scroll; title/stats leave.
// Wrapper uses display:contents in CSS so sticky is not trapped.
</script>
