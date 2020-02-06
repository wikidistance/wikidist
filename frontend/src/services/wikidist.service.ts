import { Article, ArticleNode, ArticleLink } from '../types/article';
import axios from 'axios';

const axiosInstance = axios.create({
  baseURL: 'http://160.228.22.171:8081',
});

export async function fetchArticle(
  uid: string
): Promise<{ articles: Article[]; links: ArticleLink[] }> {
  try {
    const response = await axiosInstance.request({
      method: 'POST',
      url: '/search-uid',
      data: {
        Search: uid,
        Depth: 1,
      },
    });
    const articles: Article[] = [];
    const links: ArticleLink[] = [];
    for (const article of response.data) {
      treatArticle(article, articles, links);
    }
    return { articles, links };
  } catch (e) {
    console.error(e);
    return { articles: [], links: [] };
  }
}

export async function fetchPath(
  uidSource: string,
  uidTarget: string
): Promise<{ articles: Article[]; links: ArticleLink[] }> {
  try {
    const response = await axiosInstance.request({
      method: "GET",
      url: "/shortest",
      params: {
        from: uidSource,
        to: uidTarget,
      },
    });
    const articles: Article[] = [];
    const links: ArticleLink[] = [];
    for (const article of response.data) {
      treatArticle(article, articles, links);
    }
    return { articles, links };
  } catch (e) {
    console.error(e);
    return { articles: [], links: [] };
  }
}

export async function fetchArticleByTitle(
  title: string
): Promise<{ articles: Article[]; links: ArticleLink[] }> {
  try {
    const response = await axiosInstance.request({
      method: "POST",
      url: "/search",
      data: {
        Search: title,
        Depth: 1,
      },
    });
    const articles: Article[] = [];
    const links: ArticleLink[] = [];
    for (const article of response.data) {
      treatArticle(article, articles, links);
    }
    return { articles, links };
  } catch (e) {
    console.error(e);
    return { articles: [], links: [] };
  }
}

function treatArticle(article: Article, articles: Article[], links: ArticleLink[]) {
  if (!articles.find(a => a.uid === article.uid)) {
    articles.push(article);
  }
  if (article.linked_articles !== null && article.linked_articles !== undefined) {
    for (const linkedArticle of article.linked_articles) {
      if (
        !links.find(
          l =>
            (l.source === article.uid && l.target === linkedArticle.uid) ||
            (l.target === article.uid && l.source === linkedArticle.uid)
        )
      ) {
        links.push({
          source: article.uid,
          target: linkedArticle.uid,
        });
      }
      treatArticle(linkedArticle, articles, links);
    }
  }
}
