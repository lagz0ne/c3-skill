// Mock database client (would be Prisma in real app)
export const db = {
  product: {
    async findMany() {
      return [
        { id: "1", name: "Widget", price: 9.99 },
        { id: "2", name: "Gadget", price: 19.99 },
      ];
    },
    async findUnique({ where }: { where: { id: string } }) {
      return { id: where.id, name: "Widget", price: 9.99 };
    },
  },
  order: {
    async create({ data }: { data: unknown }) {
      return { id: "order-1", ...data };
    },
  },
};
