export interface GraphState {
  pages: Map<number, Page>;
  links: Map<number, Link>;
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
