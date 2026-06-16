<template>
  <main class="auth-layout">
    <slot />
  </main>
</template>

<style scoped lang="scss">
.auth-layout {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 24px;
  background:
    linear-gradient(135deg, rgba(215, 25, 32, 0.08), transparent 34%),
    var(--braip-theme-background);
}
</style>
