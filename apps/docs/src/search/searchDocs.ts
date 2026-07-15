import {normalizeForSearch, tokenizeSearchQuery} from './normalize';
import type {SearchEntry} from './types';

const maxSearchResults = 20;

export type SearchResult = SearchEntry & {
  score: number;
  excerpt: string;
};

export type SearchResponse = {
  results: SearchResult[];
  total: number;
};

export function searchDocs(entries: SearchEntry[], query: string): SearchResponse {
  const normalizedQuery = normalizeForSearch(query);
  const tokens = tokenizeSearchQuery(query);

  if (!tokens.length) {
    return {
      results: [],
      total: 0,
    };
  }

  const matches = entries
    .map(entry => {
      const score = getScore(entry, normalizedQuery, tokens);

      if (score <= 0) return null;

      return {
        ...entry,
        score,
        excerpt: getExcerpt(entry, tokens),
      } satisfies SearchResult;
    })
    .filter((entry): entry is SearchResult => Boolean(entry))
    .sort((a, b) => b.score - a.score || a.title.localeCompare(b.title, 'fa'));

  return {
    results: matches.slice(0, maxSearchResults),
    total: matches.length,
  };
}

function getScore(
  entry: SearchEntry,
  normalizedQuery: string,
  tokens: string[],
): number {
  const fields = entry.normalized;
  const wholeDocument = [
    fields.title,
    fields.description,
    fields.headings,
    fields.content,
  ].join(' ');

  if (!tokens.every(token => wholeDocument.includes(token))) return 0;

  let score = 0;

  if (fields.title.includes(normalizedQuery)) score += 60;
  if (fields.description.includes(normalizedQuery)) score += 35;
  if (fields.headings.includes(normalizedQuery)) score += 30;
  if (fields.content.includes(normalizedQuery)) score += 12;

  for (const token of tokens) {
    if (fields.title.includes(token)) score += 20;
    if (fields.description.includes(token)) score += 12;
    if (fields.headings.includes(token)) score += 10;
    if (fields.content.includes(token)) score += 3;
  }

  return score;
}

function getExcerpt(entry: SearchEntry, tokens: string[]): string {
  const description = entry.description.trim();
  if (description && tokens.some(token => entry.normalized.description.includes(token))) {
    return description;
  }

  const normalizedContent = entry.normalized.content;
  const firstTokenIndex = tokens
    .map(token => normalizedContent.indexOf(token))
    .filter(index => index >= 0)
    .sort((a, b) => a - b)[0];

  if (firstTokenIndex === undefined) return description || entry.content.slice(0, 180);

  const source = entry.content.replace(/\s+/g, ' ').trim();
  const start = Math.max(0, firstTokenIndex - 70);
  const end = Math.min(source.length, firstTokenIndex + 170);
  const prefix = start > 0 ? '...' : '';
  const suffix = end < source.length ? '...' : '';

  return `${prefix}${source.slice(start, end).trim()}${suffix}`;
}
