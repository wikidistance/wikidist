import { GraphState } from './types';
import { Module } from 'vuex';
import { RootState } from '@/store/types';
import { actions } from './actions';
import { mutations } from './mutations';

const state: GraphState = {
  pages: [],
  links: [],
};

export const graph: Module<GraphState, RootState> = {
  actions,
  mutations,
  namespaced: true,
  state,
};
