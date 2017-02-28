import {MetricsPanelCtrl} from 'app/plugins/sdk';
import echarts from 'app/plugins/pili/echarts';
import china from './china';
import provinces from './province_in_china';
import _ from 'lodash';

echarts.registerMap('china', china);

function randomData() {
    return Math.round(Math.random() * 1000);
}

export class MapCtrl extends MetricsPanelCtrl {

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
        this._setupSeries();

        this.render();
    }

    setUnitFormat(subItem) {

    }

    changeSeriesColor(series, color) {
        series.color = color;
    }

    _setupSeries() {
        this._series = [];
        this._seriesMap = {};
        this._legendData = [];
    }

    _parseData(dataList) {
        this._series.forEach(serie => this._clearSerieData(serie.data));
        dataList.forEach(({target, datapoints}) => {
            const targetMatcher = target.match(/\{.+}/);
            if (!targetMatcher) {
                return;
            }
            let targetObj = targetMatcher[0];
            if (!targetObj) {
                return;
            }
            targetObj = this._parseTargetObjString(targetObj);
            if (provinces.indexOf(targetObj.province) == -1) {
                return;
            }
            let targetKeys = [];
            for (const key in targetObj) {
                targetKeys.push(key);
            }
            targetKeys = targetKeys.sort().filter(key => key !== 'province');
            const targetValues = targetKeys.map(key => targetObj[key]);
            const serieName = targetValues.length >0 ? targetValues.join('-')
                                                     : 'main';
            this._updateSerieData(targetObj.province, serieName, datapoints);
        });
        console.log(this._series);
        this._myChart.setOption({
            series: this._series,
            legend: {data: this._legendData},
        });
    }

    _parseTargetObjString(strObj) {
        const obj = {};
        const trim = s => s.replace(/(^\s+|\s+$)/, '');
        strObj.replace(/(\{|\})/g, '').split(',').forEach(cell => {
            const [key, value] = trim(cell).split(':').map(trim);
            obj[key] = value;
        });
        return obj;
    }

    _clearSerieData(data) {
        while(data.length > 0) {
            data.pop();
        }
        provinces.forEach(province => data.push({
            name: province, value: NaN,
        }));
    }

    _updateSerieData(province, serieName, datapoints) {
        let serie = this._seriesMap[serieName];
        if (!serie) {
            serie = {
                name: serieName,
                type: 'map',
                mapType: 'china',
                roam: false,
                data: [],
                label: {
                    normal: {
                        show: true
                    },
                    emphasis: {
                        show: true
                    }
                },
            };
            this._seriesMap[serieName] = serie;
            this._series.push(serie);
            this._legendData.push(serieName);
            this._clearSerieData(serie.data);
        }
        const cell = serie.data.find(cell => cell.name === province);
        if (cell) {
            cell.value = datapoints[0][0];
        }
    }

    onDataReceived(dataList) {
        if (!this.didRenderThisComponent) {
            this._myChart = echarts.init(document.getElementById('echart-dom'));
            // 指定图表的配置项和数据
            var option = {
                tooltip: {
                    trigger: 'item'
                },
                legend: {
                    orient: 'vertical',
                    left: 'left',
                    data: []
                },
                visualMap: {
                    min: 0,
                    max: 2500,
                    left: 'left',
                    top: 'bottom',
                    text: ['高', '低'],           // 文本，默认为数值文本
                    calculable: true
                },
                toolbox: {
                    show: true,
                    orient: 'vertical',
                    left: 'right',
                    top: 'center',
                    feature: {
                        dataView: {readOnly: false},
                        restore: {},
                        saveAsImage: {}
                    }
                },
                series: [],
            };
            this._myChart.setOption(option);
            this.didRenderThisComponent = true;
        }
        this._parseData(dataList);
    }

    onDataError() {
        console.error(`fail to load database`)
    }

    onRender() {
    }
}

MapCtrl.templateUrl = 'module.html';
