import type { CollectionEntry } from "astro:content";

export type BlogPost = CollectionEntry<"blog">;

export function isPublished(post: BlogPost) {
  return import.meta.env.PROD ? post.data.draft !== true : true;
}

export function sortPosts(posts: BlogPost[]) {
  return [...posts].sort(
    (left, right) => right.data.pubDate.getTime() - left.data.pubDate.getTime(),
  );
}

export function getPostSlug(post: BlogPost) {
  return post.data.slug || post.id;
}

export function getPostUrl(post: BlogPost) {
  return `/blog/${getPostSlug(post)}`;
}

export function getPostCategory(post: BlogPost) {
  return post.data.category || post.data.tags[0] || "General";
}

export function slugifySegment(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/&/g, " and ")
    .replace(/[^a-z0-9\u0600-\u06FF]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

export function getCategorySlug(category: string) {
  return slugifySegment(category);
}

export function getCategoryUrl(category: string) {
  return `/blog/category/${getCategorySlug(category)}`;
}

export function getCategoryLabelFromSlug(posts: BlogPost[], slug: string) {
  const match = posts.find((post) => getCategorySlug(getPostCategory(post)) === slug);
  return match ? getPostCategory(match) : slug;
}

export function groupPostsByCategory(posts: BlogPost[]) {
  const grouped = new Map<string, BlogPost[]>();

  for (const post of posts) {
    const category = getPostCategory(post);
    const current = grouped.get(category) || [];
    current.push(post);
    grouped.set(category, current);
  }

  return [...grouped.entries()]
    .map(([category, categoryPosts]) => ({
      category,
      slug: getCategorySlug(category),
      posts: sortPosts(categoryPosts),
    }))
    .sort((left, right) => right.posts.length - left.posts.length);
}
