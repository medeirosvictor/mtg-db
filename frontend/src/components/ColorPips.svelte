<script lang="ts">
  import { parseColors, getColorHex, type MtgColor } from '../lib/colors';

  export let colors: string;

  $: parsed = parseColors(colors);

  const pipStyle = (color: MtgColor) => {
    const hex = getColorHex(color);
    const border = color === 'W' ? '#ccc' : color === 'B' ? '#555' : hex;
    return `background: ${hex}; border-color: ${border};`;
  };
</script>

<div class="color-pips">
  {#each parsed as color}
    <span class="pip" style={pipStyle(color)} title={color}></span>
  {/each}
</div>

<style>
  .color-pips {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  .pip {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    border: 1.5px solid;
    display: inline-block;
    flex-shrink: 0;
  }
</style>
