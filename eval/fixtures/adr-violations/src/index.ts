import { Hono } from 'hono'
import { authMiddleware } from './middleware/auth'
import { taskRoutes } from './routes/tasks'
import { userRoutes } from './routes/users'

const app = new Hono()

app.use('/*', authMiddleware)
app.route('/api/tasks', taskRoutes)
app.route('/api/users', userRoutes)

export default app
