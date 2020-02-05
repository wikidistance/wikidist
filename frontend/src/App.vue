<template>
  <div id="app">
    <div id="top-bar">
      <h1>WIKIDIST</h1>
    </div>
    <Menu @action="onAction" />
    <Graph :loading="loading" :links="links" :articles="articles" />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator';
import Graph from './components/Graph.vue';
import Menu from './components/Menu.vue';
import { ArticleNode, ArticleLink, Article } from './types/article';
import { fetchArticle, fetchArticleByTitle, fetchPath } from '@/services/wikidist.service';

@Component({
  components: {
    Graph,
    Menu,
  },
})
export default class App extends Vue {
  private loading: boolean = false;
  private articles: Article[] = [];
  private links: ArticleLink[] = [];

  public async onAction(args: { source: string; target: string }) {
    this.loading = true;
    const source = await fetchArticleByTitle(args.source);
    const target = await fetchArticleByTitle(args.target);
    if (source.articles.length > 0 && target.articles.length > 0) {
      const articles = await fetchPath(source.articles[0].uid, target.articles[0].uid);
      this.articles = articles.articles;
      this.links = articles.links;
    }
    this.loading = false;
  }
}
</script>

<style>
#top-bar {
  height: 50px;
  width: 100%;
  background-color: #2b7a8c;
  color: white;
  padding: 10px;
  box-sizing: border-box;
}

#app {
  font-family: 'Ubuntu', sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
  height: 100%;
}

html,
body {
  margin: 0;
  padding: 0;
  height: 100%;
  box-sizing: border-box;
}

h1,
h2,
h3,
h4,
h5,
h6,
p {
  padding: 0;
  margin: 0;
}

input,
button {
  outline: none;
  border: none;
}

button {
  cursor: pointer;
}

input {
  cursor: text;
}
</style>
