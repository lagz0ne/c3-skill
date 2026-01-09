"use client";

import Link from "next/link";

interface Post {
  id: number;
  title: string;
  content: string | null;
  createdAt: Date;
  author: { name: string | null; image: string | null };
}

export function PostCard({ post }: { post: Post }) {
  return (
    <Link href={`/post/${post.id}`}>
      <article className="p-6 bg-slate-800 rounded-lg border border-slate-700 hover:border-slate-600 transition-colors">
        <h2 className="text-xl font-semibold text-white mb-2">{post.title}</h2>
        {post.content && (
          <p className="text-gray-400 mb-4 line-clamp-3">{post.content}</p>
        )}
        <div className="flex items-center gap-2 text-sm text-gray-500">
          {post.author.image && (
            <img
              src={post.author.image}
              alt=""
              className="w-6 h-6 rounded-full"
            />
          )}
          <span>{post.author.name ?? "Anonymous"}</span>
          <span>•</span>
          <time>{new Date(post.createdAt).toLocaleDateString()}</time>
        </div>
      </article>
    </Link>
  );
}
