import { GraphState } from './types';
import { ActionTree } from 'vuex';
import { RootState } from '@/store/types';
import { getPage } from '@/services/wikidist';

export const actions: ActionTree<GraphState, RootState> = {
  async fetchPage({ commit }, pageName: string) {
    const result = await getPage(pageName);
    if (result.center !== undefined) {
      for (const page of result.pages) {
        commit('addPage', page);
      }
      for (const link of result.links) {
        commit('addLink', link);
      }
    }
  },
};
