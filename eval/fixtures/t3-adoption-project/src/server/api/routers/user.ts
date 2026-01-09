import { z } from "zod";
import { createTRPCRouter, protectedProcedure, publicProcedure } from "@/server/api/trpc";

export const userRouter = createTRPCRouter({
  getProfile: publicProcedure
    .input(z.object({ id: z.string() }))
    .query(async ({ ctx, input }) => {
      return ctx.db.user.findUnique({
        where: { id: input.id },
        select: {
          id: true,
          name: true,
          image: true,
          posts: {
            where: { published: true },
            select: { id: true, title: true, createdAt: true },
          },
        },
      });
    }),

  updateProfile: protectedProcedure
    .input(z.object({ name: z.string().min(1).optional() }))
    .mutation(async ({ ctx, input }) => {
      return ctx.db.user.update({
        where: { id: ctx.session.user.id },
        data: { name: input.name },
      });
    }),

  getMe: protectedProcedure.query(async ({ ctx }) => {
    return ctx.db.user.findUnique({
      where: { id: ctx.session.user.id },
      include: {
        posts: { orderBy: { createdAt: "desc" } },
      },
    });
  }),
});
