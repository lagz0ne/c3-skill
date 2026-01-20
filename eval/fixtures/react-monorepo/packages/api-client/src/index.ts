const API_URL = process.env.API_URL || "http://localhost:3001";

export const api = {
  products: {
    async list() {
      const res = await fetch(`${API_URL}/products`);
      return res.json();
    },
    async get(id: string) {
      const res = await fetch(`${API_URL}/products/${id}`);
      return res.json();
    },
  },
  orders: {
    async create(items: { productId: string; quantity: number }[]) {
      const res = await fetch(`${API_URL}/orders`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ items }),
      });
      return res.json();
    },
  },
};
