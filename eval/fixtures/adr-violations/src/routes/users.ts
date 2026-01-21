import { Hono } from 'hono'

export const userRoutes = new Hono()

userRoutes.get('/profile', (c) => c.json({ user: {} }))
