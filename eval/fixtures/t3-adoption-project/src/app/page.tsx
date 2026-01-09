"use client";

import { api } from "@/lib/trpc";
import { PostCard } from "@/components/post-card";
import { Navbar } from "@/components/navbar";

export default function HomePage() {
  const { data: posts, isLoading } = api.post.getAll.useQuery();

  return (
    <div className="container mx-auto px-4 py-8">
      <Navbar />
      <div className="mt-8">
        <h1 className="text-4xl font-bold text-white mb-8">Latest Posts</h1>
        {isLoading ? (
          <div className="text-gray-400">Loading posts...</div>
        ) : posts?.length === 0 ? (
          <div className="text-gray-400">No posts yet. Be the first to create one!</div>
        ) : (
          <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
            {posts?.map((post) => (
              <PostCard key={post.id} post={post} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
