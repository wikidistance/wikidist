import { GraphState, Page, Link } from './types';
import { MutationTree } from 'vuex';

export const mutations: MutationTree<GraphState> = {
  addPage(state, page: Page) {
    if (!state.pages.has(page.id)) {
      state.pages.set(page.id, page);
    }
  },
  setPages(state, pages: Page[]) {
    state.pages = new Map(pages.map(p => [p.id, p]));
  },
  addLink(state, link: Link) {
    if (!state.links.has(link.id)) {
      state.links.set(link.id, link);
    }
  },
  setLinks(state, links: Link[]) {
    state.links = new Map(links.map(l => [l.id, l]));
  },
};
