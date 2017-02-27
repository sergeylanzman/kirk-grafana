import {MetricsPanelCtrl} from 'app/plugins/sdk';
import echarts from 'app/plugins/pili/echarts';
import _ from 'lodash';

export class PieChartCtrl extends MetricsPanelCtrl {

    constructor($scope, $injector, $rootScope) {
        super($scope, $injector);
        this.$rootScope = $rootScope;

        var panelDefaults = {
            pieType: 'pie',
            legend: {
                show: true, // disable/enable legend
                values: true
            },
            links: [],
            datasource: null,
            maxDataPoints: 3,
            interval: null,
            targets: [{}],
            cacheTimeout: null,
            nullPointMode: 'connected',
            legendType: 'Under graph',
            aliasColors: {},
            format: 'short',
            valueName: 'current',
            strokeWidth: 1,
            fontSize: '80%',
            combine: {
                threshold: 0.0,
                label: 'Others'
            }
        };

        _.defaults(this.panel, panelDefaults);
        _.defaults(this.panel.legend, panelDefaults.legend);

        this.events.on('render', this.onRender.bind(this));
        this.events.on('data-received', this.onDataReceived.bind(this));
        this.events.on('data-snapshot-load', this.onDataReceived.bind(this));
        this.events.on('data-error', this.onDataError.bind(this));
        // this.events.on('init-edit-mode', this.onInitEditMode.bind(this));

        this.didRenderThisComponent = false;

        this.render();
    }

    setUnitFormat(subItem) {

    }

    changeSeriesColor(series, color) {
        series.color = color;
    }

    onDataReceived(dataList) {
        this.htmlContent = JSON.stringify(dataList);

      if (!this.didRenderThisComponent) {

        console.log(echarts);
        var myChart = echarts.init(document.getElementById('echart-dom'));

                // 指定图表的配置项和数据
        var option = {
            title: {
                text: 'ECharts 入门示例'
            },
            tooltip: {},
            legend: {
                data:['销量']
            },
            xAxis: {
                data: ["衬衫","羊毛衫","雪纺衫","裤子","高跟鞋","袜子"]
            },
            yAxis: {},
            series: [{
                name: '销量',
                type: 'bar',
                data: [5, 20, 36, 10, 10, 20]
            }]
        };
        myChart.setOption(option);

        this.didRenderThisComponent = true;
      }
    }

    onDataError() {
        console.error(`fail to load database`)
    }

    onRender() {
    }
}

PieChartCtrl.templateUrl = 'module.html';
