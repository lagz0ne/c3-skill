import { createCallerFactory, createTRPCRouter } from "@/server/api/trpc";
import { postRouter } from "@/server/api/routers/post";
import { userRouter } from "@/server/api/routers/user";

export const appRouter = createTRPCRouter({
  post: postRouter,
  user: userRouter,
});

export type AppRouter = typeof appRouter;
export const createCaller = createCallerFactory(appRouter);
