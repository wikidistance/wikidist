import { GraphState, Page, Link } from './types';
import { MutationTree } from 'vuex';

export const mutations: MutationTree<GraphState> = {
  addPage(state, page: Page) {
    if (state.pages.find(p => p.id === page.id) === undefined) {
      state.pages.push(page);
    }
  },
  setPages(state, pages: Page[]) {
    state.pages = pages;
  },
  addLink(state, link: Link) {
    if (state.links.find(l => l.id === link.id) === undefined) {
      state.links.push(link);
    }
  },
  setLinks(state, links: Link[]) {
    state.links = links;
  },
};
