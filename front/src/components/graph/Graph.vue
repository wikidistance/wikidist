<template>
  <div>
    <svg height="500" width="500">
      <g transform="translate(250 250)" v-if="tree !== undefined">
        <path
          class="link"
          v-for="link of tree.links()"
          :key="`${link.source.data.id}-${link.target.data.id}`"
          :d="linkPath(link)"
        />
        <g
          v-for="node of tree.descendants()"
          :key="node.data.id"
          :transform="nodeTransform(node)"
        >
          <circle class="circle" />
          <text x="15" y="5" class="text">{{ node.data.name }}</text>
        </g>
      </g>
    </svg>
  </div>
</template>
<script lang="ts">
import { Vue, Component, Prop, Watch } from 'vue-property-decorator';
import { Link, Page } from '../../store/modules/graph/types';
import {
  hierarchy,
  tree,
  HierarchyPointNode,
  HierarchyPointLink,
} from 'd3-hierarchy';

@Component
export default class Graph extends Vue {
  @Prop() public links!: Link[];
  @Prop() public pages!: Page[];
  @Prop() public root!: Page;

  private static project(t: number, r: number): { x: number; y: number } {
    return {
      x: Math.cos(t) * r,
      y: Math.sin(t) * r,
    };
  }

  public nodeTransform(node: HierarchyPointNode<Page>): string {
    const projection = Graph.project(node.x, node.y);
    return `translate(${projection.x} ${projection.y})`;
  }

  public linkPath(link: HierarchyPointLink<Page>): string {
    const sourceProjection = Graph.project(link.source.x, link.source.y);
    const targetProjection = Graph.project(link.target.x, link.target.y);
    return `M ${sourceProjection.x} ${sourceProjection.y} L ${targetProjection.x} ${targetProjection.y}`;
  }

  public get tree(): HierarchyPointNode<Page> | undefined {
    if (this.pages.length > 0 && this.root !== undefined) {
      const h = hierarchy(this.root, (p: Page): Page[] | null => {
        const children: Page[] = this.links
          .filter(l => l.from === p.id)
          .map(l => this.pages.find(p => p.id === l.to))
          .filter(p => p !== undefined) as Page[];
        return children.length > 0 ? children : null;
      });
      return tree().size([2 * Math.PI, 100])(h) as HierarchyPointNode<Page>;
    }
    return undefined;
  }
}
</script>
<style scoped>
.circle {
  r: 10;
}

.link {
  stroke: black;
  stroke-width: 3;
}

.text {
  fill: red;
}
</style>
