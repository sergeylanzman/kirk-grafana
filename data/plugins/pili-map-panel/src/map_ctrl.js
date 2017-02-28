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

    _parseData(dataList, option) {
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
        option.series = this._series;
        option.legend = option.legend || {data: this._legendData};
        return option;
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
        provinces.forEach(province => {
            const cell = data.find(cell => cell.name === province);
            if (cell) {
                cell.value = NaN;
            } else {
                data.push({
                    name: province,
                    value: NaN,
                });
            }
        });
    }

    _updateSerieData(province, serieName, datapoints) {
        let serie = this._seriesMap[serieName];
        if (!serie) {
            serie = {
                name: serieName,
                type: 'map',
                mapType: 'china',
                roam: false,
                showLegendSymbol: true,
                data: [],
                itemStyle: {
                    normal: {
                        label: {show: false},
                        color: "#32cd32",
                    },
                    emphasis: {
                        borderWidth: 1,
                        borderColor: "#fff",
                        color: "#32cd32",
                    },
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
        let option = {};
        if (!this.didRenderThisComponent) {
            this._myChart = echarts.init(document.getElementById('echart-dom'));
            // 指定图表的配置项和数据
            option = {
                tooltip: {
                    trigger: 'item',
                    formatter: '{b}<br/>{c}',
                },
                legend: {
                    show: true,
                    orient: "vertical",
                    left: "left",
                    data: this.legendData,
                },
                visualMap: {
                    left: "left",
                    top: "bottom",
                    text: ["高", "低"],
                    calculable: true,
                    max: 20,
                    min: 10,
                    inRange: {
                        color: [
                            "#0C8918",
                            "#9ED900",
                            "#F05654",
                            "#C91F37",
                        ],
                    },
                },
                toolbox: {
                    show: true,
                    orient: "vertical",
                    left: "right",
                    top: "center",
                },
                series: [],
            };
            this.didRenderThisComponent = true;
        }
        option = this._parseData(dataList, option);
        this._myChart.setOption(option);
    }

    onDataError() {
        console.error(`fail to load database`)
    }

    onRender() {
    }
}

MapCtrl.templateUrl = 'module.html';
