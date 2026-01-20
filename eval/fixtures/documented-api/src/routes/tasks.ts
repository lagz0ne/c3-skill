import { Hono } from "hono";
import { db } from "../db";

export const taskRoutes = new Hono();

taskRoutes.get("/", async (c) => {
  const tasks = await db.query("SELECT * FROM tasks WHERE user_id = ?", [c.get("userId")]);
  return c.json(tasks);
});

taskRoutes.post("/", async (c) => {
  const { title, description } = await c.req.json();
  const result = await db.query(
    "INSERT INTO tasks (title, description, user_id) VALUES (?, ?, ?)",
    [title, description, c.get("userId")]
  );
  return c.json({ id: result.insertId, title, description });
});

taskRoutes.patch("/:id", async (c) => {
  const id = c.req.param("id");
  const updates = await c.req.json();
  await db.query("UPDATE tasks SET ? WHERE id = ? AND user_id = ?", [updates, id, c.get("userId")]);
  return c.json({ success: true });
});

taskRoutes.delete("/:id", async (c) => {
  const id = c.req.param("id");
  await db.query("DELETE FROM tasks WHERE id = ? AND user_id = ?", [id, c.get("userId")]);
  return c.json({ success: true });
});
