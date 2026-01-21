import type { Context, Next } from 'hono'

export async function authMiddleware(c: Context, next: Next) {
  const token = c.req.header('Authorization')?.replace('Bearer ', '')
  if (!token) {
    return c.json({ error: 'Unauthorized' }, 401)
  }
  // JWT verification would happen here
  c.set('userId', 'user-123')
  await next()
}
