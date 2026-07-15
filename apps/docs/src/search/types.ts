export type SearchEntry = {
  id: string;
  title: string;
  path: string;
  description: string;
  headings: string[];
  content: string;
  normalized: {
    title: string;
    description: string;
    headings: string;
    content: string;
  };
};

export type SearchData = {
  entries: SearchEntry[];
};

