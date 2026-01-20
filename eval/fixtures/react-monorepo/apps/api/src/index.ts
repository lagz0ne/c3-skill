import { Hono } from "hono";
import { cors } from "hono/cors";
import { db } from "@acme/database";

const app = new Hono();

app.use("/*", cors());

app.get("/products", async (c) => {
  const products = await db.product.findMany();
  return c.json(products);
});

app.get("/products/:id", async (c) => {
  const product = await db.product.findUnique({ where: { id: c.req.param("id") } });
  return c.json(product);
});

app.post("/orders", async (c) => {
  const { items } = await c.req.json();
  const order = await db.order.create({ data: { items } });
  return c.json(order);
});

export default app;
