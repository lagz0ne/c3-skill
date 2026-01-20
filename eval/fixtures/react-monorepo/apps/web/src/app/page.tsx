import { Button } from "@acme/ui";
import { api } from "@acme/api-client";

export default async function Home() {
  const products = await api.products.list();

  return (
    <main>
      <h1>Acme Store</h1>
      <div className="grid">
        {products.map((p) => (
          <div key={p.id}>
            <h2>{p.name}</h2>
            <p>{p.price}</p>
            <Button>Add to Cart</Button>
          </div>
        ))}
      </div>
    </main>
  );
}
