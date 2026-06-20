import { onBeforeUnmount, onMounted, ref, watch, type Ref } from 'vue'
import * as echarts from 'echarts/core'
import { LineChart, BarChart, PieChart } from 'echarts/charts'
import {
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  TitleComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

echarts.use([
  LineChart,
  BarChart,
  PieChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  TitleComponent,
  CanvasRenderer,
])

export type EChartsOption = echarts.EChartsCoreOption

export function useECharts(optionRef: Ref<EChartsOption>) {
  const el = ref<HTMLDivElement | null>(null)
  let chart: echarts.ECharts | null = null

  const resize = () => chart?.resize()

  onMounted(() => {
    if (el.value) {
      chart = echarts.init(el.value, 'dark')
      chart.setOption(optionRef.value)
      window.addEventListener('resize', resize)
    }
  })

  watch(optionRef, (opt) => chart?.setOption(opt, true), { deep: true })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', resize)
    chart?.dispose()
    chart = null
  })

  return { el }
}
