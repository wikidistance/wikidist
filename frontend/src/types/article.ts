import { SimulationNodeDatum, SimulationLinkDatum } from 'd3-force';

export interface Article {
  uid: string;
  url: string;
  title: string;
  linked_articles: Article[] | null;
}

export interface ArticleNode extends SimulationNodeDatum {
  article: Article;
}

export interface ArticleLink extends SimulationLinkDatum<ArticleNode> {
  source: string;
  target: string;
}
