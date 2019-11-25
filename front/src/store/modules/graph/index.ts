import { GraphState } from './types';
import { Module } from 'vuex';
import { RootState } from '@/store/types';

const state: GraphState = {
  pages: new Map(),
  links: new Map(),
};

export const graph: Module<GraphState, RootState> = {
  namespaced: true,
  state,
};
