export interface GraphState {
  pages: Page[];
  links: Link[];
}

export interface Page {
  id: number;
  name: string;
}

export interface Link {
  id: number;
  from: number;
  to: number;
}
