import { z } from "zod";
import { createTRPCRouter, protectedProcedure, publicProcedure } from "@/server/api/trpc";

export const postRouter = createTRPCRouter({
  getAll: publicProcedure.query(async ({ ctx }) => {
    return ctx.db.post.findMany({
      where: { published: true },
      include: { author: { select: { name: true, image: true } } },
      orderBy: { createdAt: "desc" },
    });
  }),

  getById: publicProcedure
    .input(z.object({ id: z.number() }))
    .query(async ({ ctx, input }) => {
      return ctx.db.post.findUnique({
        where: { id: input.id },
        include: { author: { select: { name: true, image: true } } },
      });
    }),

  create: protectedProcedure
    .input(z.object({ title: z.string().min(1), content: z.string().optional() }))
    .mutation(async ({ ctx, input }) => {
      return ctx.db.post.create({
        data: {
          title: input.title,
          content: input.content,
          authorId: ctx.session.user.id,
        },
      });
    }),

  update: protectedProcedure
    .input(z.object({
      id: z.number(),
      title: z.string().min(1).optional(),
      content: z.string().optional(),
      published: z.boolean().optional(),
    }))
    .mutation(async ({ ctx, input }) => {
      const post = await ctx.db.post.findUnique({ where: { id: input.id } });
      if (!post || post.authorId !== ctx.session.user.id) {
        throw new Error("Not authorized");
      }
      return ctx.db.post.update({
        where: { id: input.id },
        data: {
          title: input.title,
          content: input.content,
          published: input.published,
        },
      });
    }),

  delete: protectedProcedure
    .input(z.object({ id: z.number() }))
    .mutation(async ({ ctx, input }) => {
      const post = await ctx.db.post.findUnique({ where: { id: input.id } });
      if (!post || post.authorId !== ctx.session.user.id) {
        throw new Error("Not authorized");
      }
      return ctx.db.post.delete({ where: { id: input.id } });
    }),

  getMyPosts: protectedProcedure.query(async ({ ctx }) => {
    return ctx.db.post.findMany({
      where: { authorId: ctx.session.user.id },
      orderBy: { createdAt: "desc" },
    });
  }),
});
