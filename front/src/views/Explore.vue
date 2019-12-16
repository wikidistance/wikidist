<template>
  <div>
    <Graph :pages="pages" :links="links" :root="pages[0]" />
  </div>
</template>
<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import Graph from '@/components/graph/Graph.vue';

@Component({
  components: {
    Graph,
  },
})
export default class Explore extends Vue {
  public get pages() {
    return this.$store.state.graph.pages;
  }

  public get links() {
    return this.$store.state.graph.links;
  }

  public mounted() {
    setTimeout(() => this.$store.dispatch('graph/fetchPage', 'France'), 1000);
  }
}
</script>
