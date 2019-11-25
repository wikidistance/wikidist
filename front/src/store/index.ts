import Vuex from 'vuex';
import { RootState } from './types';
import Vue from 'vue';
import { graph } from './modules/graph/index';

Vue.use(Vuex);

export const store = new Vuex.Store<RootState>({
  modules: {
    graph,
  },
  state: {},
  strict: true,
});
