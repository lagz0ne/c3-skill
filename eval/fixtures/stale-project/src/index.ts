// v2: Consolidated into single file
import { Hono } from "hono";

const app = new Hono();

// Posts API
app.get("/posts", (c) => c.json([]));
app.post("/posts", (c) => c.json({ id: 1 }));
app.get("/posts/:id", (c) => c.json({ id: c.req.param("id") }));

// Comments API (was separate, now inline)
app.get("/posts/:id/comments", (c) => c.json([]));
app.post("/posts/:id/comments", (c) => c.json({ id: 1 }));

// Auth (simplified, was middleware)
app.post("/login", (c) => c.json({ token: "xxx" }));

export default app;
