import { Hono } from 'hono'

export const taskRoutes = new Hono()

taskRoutes.get('/', (c) => c.json({ tasks: [] }))
taskRoutes.post('/', (c) => c.json({ created: true }))
