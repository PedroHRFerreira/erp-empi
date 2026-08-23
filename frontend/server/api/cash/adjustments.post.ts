export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig(event);
  const authorization = getHeader(event, "authorization");
  return $fetch<unknown>(`${config.apiBase}/api/cash/adjustments`, {
    method: "POST",
    body: await readBody(event),
    headers: authorization ? { Authorization: authorization } : undefined,
  });
});
