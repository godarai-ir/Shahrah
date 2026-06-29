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
  return post.data.category || post.data.tags[0] || "عمومی";
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

export function getTagSlug(tag: string) {
  return slugifySegment(tag);
}

export function getTagUrl(tag: string) {
  return `/blog/tag/${getTagSlug(tag)}`;
}

export function getTagLabelFromSlug(posts: BlogPost[], slug: string) {
  const match = posts.find((post) => post.data.tags.some((tag) => getTagSlug(tag) === slug));
  return match ? match.data.tags.find((tag) => getTagSlug(tag) === slug) || slug : slug;
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

export function groupPostsByTag(posts: BlogPost[]) {
  const grouped = new Map<string, BlogPost[]>();

  for (const post of posts) {
    for (const tag of post.data.tags) {
      const current = grouped.get(tag) || [];
      current.push(post);
      grouped.set(tag, current);
    }
  }

  return [...grouped.entries()]
    .map(([tag, taggedPosts]) => ({
      tag,
      slug: getTagSlug(tag),
      posts: sortPosts(taggedPosts),
    }))
    .sort((left, right) => right.posts.length - left.posts.length);
}

export function formatPersianDate(date: Date) {
  return date.toLocaleDateString("fa-IR", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

export function getReadingTime(post: BlogPost) {
  if (post.data.readingTime) {
    return post.data.readingTime;
  }

  const words = (post.body || "").split(/\s+/u).filter(Boolean).length;
  return Math.max(1, Math.round(words / 220));
}

export function getRelatedPosts(posts: BlogPost[], currentPost: BlogPost, limit = 3) {
  const currentCategory = getPostCategory(currentPost);
  const currentTags = new Set(currentPost.data.tags);

  return posts
    .filter((post) => post.id !== currentPost.id)
    .map((post) => {
      let score = 0;
      if (getPostCategory(post) === currentCategory) score += 3;
      for (const tag of post.data.tags) {
        if (currentTags.has(tag)) score += 1;
      }

      return { post, score };
    })
    .sort(
      (left, right) =>
        right.score - left.score ||
        right.post.data.pubDate.getTime() - left.post.data.pubDate.getTime(),
    )
    .slice(0, limit)
    .map((item) => item.post);
}

export function paginatePosts(posts: BlogPost[], currentPage: number, pageSize: number) {
  const totalPages = Math.max(1, Math.ceil(posts.length / pageSize));
  const safePage = Math.min(Math.max(currentPage, 1), totalPages);
  const start = (safePage - 1) * pageSize;

  return {
    currentPage: safePage,
    totalPages,
    pageSize,
    items: posts.slice(start, start + pageSize),
  };
}

export function getArchivePageUrl(page: number) {
  return page <= 1 ? "/blog" : `/blog/page/${page}`;
}
