<template>
  <div class="container">
    <label class="label">{{ label }}</label>
    <input :value="value" class="search" @input="input" />
  </div>
</template>
<script lang="ts">
import { Component, Prop, Vue, Emit } from 'vue-property-decorator';

@Component
export default class SearchBar extends Vue {
  @Prop() public value!: string;
  @Prop() public label!: string;

  @Emit()
  public input(e: Event) {
    return (e.target as HTMLInputElement).value;
  }
}
</script>
<style scoped>
.container {
  display: flex;
  flex-direction: column;
}
.search {
  padding: 10px 15px;
  border-radius: 40px;
}
.label {
  color: white;
  padding: 0 0 5px 15px;
  text-align: left;
}
</style>
