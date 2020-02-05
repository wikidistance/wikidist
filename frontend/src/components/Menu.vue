<template>
  <div id="menu" :class="{ retracted }">
    <button id="menu-button" @click="retracted = !retracted">
      <font-awesome-icon color="white" :icon="retracted ? 'caret-right' : 'caret-left'" />
    </button>
    <SearchBar v-model="source" label="Source" />
    <br />
    <SearchBar v-model="target" label="Target" />
    <br />
    <button id="action-button" :disabled="disabled" @click="action">GO</button>
  </div>
</template>
<script lang="ts">
import { Component, Prop, Vue, Emit } from 'vue-property-decorator';
import SearchBar from '@/components/SearchBar.vue';

@Component({
  components: {
    SearchBar,
  },
})
export default class Menu extends Vue {
  private source: string = '';
  private target: string = '';
  private retracted: boolean = false;

  @Emit()
  public action() {
    return {
      source: this.source,
      target: this.target,
    };
  }

  public get disabled() {
    return this.source.length == 0 || this.target.length == 0;
  }
}
</script>
<style scoped>
#menu {
  position: absolute;
  width: max-content;
  text-align: center;
  padding: 20px;
  background-color: #75b1bfaa;
  animation-name: slide-in;
  animation-duration: 0.5s;
  z-index: 1;
}

#menu.retracted {
  animation-name: slide-out;
  animation-duration: 0.5s;
  transform: translate(-100%);
}

@keyframes slide-out {
  from {
    transform: translate(0);
  }
  to {
    transform: translate(-100%);
  }
}

@keyframes slide-in {
  from {
    transform: translate(-100%);
  }
  to {
    transform: translate(0);
  }
}

#menu-button {
  position: absolute;
  height: 50px;
  width: 25px;
  background-color: #f47370;
  left: 100%;
  top: 50%;
  transform: translate(-25%, -50%);
  font-size: 1.2em;
  padding-left: 12px;
}

#action-button {
  padding: 5px 15px;
  border-radius: 40px;
  background-color: #f47370;
  color: white;
}

#menu-button:hover,
#action-button:hover {
  background-color: #c75d5b;
}

#action-button:active,
#menu-button:active {
  background-color: #a34d4b;
}
</style>
