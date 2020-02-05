<template>
  <div id="graph" ref="svg">
    <svg v-show="!loading" height="100%" width="100%">
      <g :transform="centerTranslate">
        <line
          v-for="(link, index) of links"
          :key="index"
          stroke="black"
          stroke-width="2"
          :x1="nodes[link.source.index].x"
          :y1="nodes[link.source.index].y"
          :x2="nodes[link.target.index].x"
          :y2="nodes[link.target.index].y"
        />
        <g
          v-for="node of nodes"
          :key="node.article.uid"
          :transform="`translate(${node.x} ${node.y})`"
        >
          <circle r="10" fill="red"></circle>
          <text x="15" font-size="0.8em">{{ node.article.title }}</text>
        </g>
      </g>
    </svg>
    <img v-show="loading" id="loader" src="@/assets/loader.gif" />
  </div>
</template>
<script lang="ts">
import { Component, Prop, Vue, Watch } from 'vue-property-decorator';
import { forceManyBody, forceSimulation, Simulation, forceLink, forceCenter } from 'd3-force';
import { ArticleNode, ArticleLink, Article } from '@/types/article';

@Component
export default class Graph extends Vue {
  @Prop() loading!: boolean;
  @Prop() links!: ArticleLink[];
  @Prop() articles!: Article[];

  private height: number = 0;
  private width: number = 0;

  private nodes: ArticleNode[] = [];

  mounted() {
    this.height = (this.$refs.svg as Element).clientHeight;
    this.width = (this.$refs.svg as Element).clientWidth;
  }

  updated() {
    for (const article of this.articles) {
      if (!this.nodes.find(n => n.article.uid === article.uid)) {
        this.nodes.push({
          article,
          x: 0,
          y: 0,
        });
      }
    }
    this.simulation.nodes(this.nodes).force('charge', forceManyBody())
    .force(
      'links',
      forceLink(this.links)
        .id(node => (node as ArticleNode).article.uid)
        .distance(100)
        .strength(1)
    )
    .force('center', forceCenter()).restart();
  }

  public get centerTranslate() {
    return `translate(${this.width / 2} ${this.height / 2})`;
  }

  private simulation: Simulation<ArticleNode, ArticleLink> = forceSimulation(this.nodes)
    .force('charge', forceManyBody())
    .force(
      'links',
      forceLink(this.links)
        .id(node => (node as ArticleNode).article.uid)
        .distance(100)
        .strength(1)
    )
    .force('center', forceCenter());
}
</script>
<style>
#loader {
  height: 300px;
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
}

#graph {
  height: calc(100% - 50px);
  display: flex;
}
</style>
