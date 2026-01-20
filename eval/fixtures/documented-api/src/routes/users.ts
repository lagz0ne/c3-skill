import { Hono } from "hono";
import { db } from "../db";

export const userRoutes = new Hono();

userRoutes.get("/me", async (c) => {
  const user = await db.query("SELECT id, email, name FROM users WHERE id = ?", [c.get("userId")]);
  return c.json(user[0]);
});

userRoutes.patch("/me", async (c) => {
  const updates = await c.req.json();
  await db.query("UPDATE users SET ? WHERE id = ?", [updates, c.get("userId")]);
  return c.json({ success: true });
});
