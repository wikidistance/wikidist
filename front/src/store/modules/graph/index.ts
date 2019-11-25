import { GraphState } from './types';
import { Module } from 'vuex';
import { RootState } from '@/store/types';

const state: GraphState = {
  pages: [],
  links: [],
};

export const graph: Module<GraphState, RootState> = {
  namespaced: true,
  state,
};
