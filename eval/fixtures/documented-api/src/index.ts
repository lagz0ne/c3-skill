import { Hono } from "hono";
import { taskRoutes } from "./routes/tasks";
import { userRoutes } from "./routes/users";
import { authMiddleware } from "./middleware/auth";

const app = new Hono();

app.use("/api/*", authMiddleware);
app.route("/api/tasks", taskRoutes);
app.route("/api/users", userRoutes);

export default app;
