import {MetricsPanelCtrl} from 'app/plugins/sdk';
import echarts from 'app/plugins/pili/echarts';
import china from './china';
import './dark_theme';
import provinces from './province_in_china';
import _ from 'lodash';

echarts.registerMap('china', china);

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
        this._legend = {};
        this._yAxis = {};
    }

    _parseData(dataList, option) {
        let provinceData = [];
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

            const province = targetObj.province;
            const value = datapoints[0][0];

            this._updateSerieData(province, serieName, value);
            provinceData.push({province, value});
        });
        provinceData = provinceData.sort((element1, element2) => element2.value - element1.value);
        while (provinceData.length > 10) {
            provinceData.pop();
        }
        this._yAxis.data = provinceData.reverse().map(({province, value}) => `${province} : ${value}`);
        this._legend.show = this._legend.data.length > 1;

        option.series = this._series;
        option.legend = this._legend;
        option.yAxis = this._yAxis;
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

    _updateSerieData(province, serieName, value) {
        let serie = this._seriesMap[serieName];
        if (!serie) {
            serie = {
                name: serieName,
                type: 'map',
                mapType: 'china',
                roam: false,
                showLegendSymbol: false,
                data: [],
                itemStyle: {
                    normal: {
                        label: {show: false},
                    },
                    emphasis: {
                        borderWidth: 1,
                        borderColor: "#fff",
                    },
                },
            };
            this._seriesMap[serieName] = serie;
            this._series.push(serie);
            this._legend.data.push(serieName);
            this._clearSerieData(serie.data);
        }
        const cell = serie.data.find(cell => cell.name === province);
        if (cell) {
            cell.value = value;
        }
    }

    onDataReceived(dataList) {
        let option = {};
        if (!this.didRenderThisComponent) {
            this._myChart = echarts.init(document.getElementById('echart-dom'), 'dark');
            this._yAxis = {
                type: 'category',
                nameGap: 16,
                position: 'right',
                data: [],
            };
            this._legend = {
                show: true,
                orient: "vertical",
                left: "left",
                data: [],
            };
            option = {
                legend: this._legend,
                yAxis: this._yAxis,
                series: [],

                tooltip: {
                    trigger: 'item',
                    formatter: '{b}<br/>{c}',
                },
                visualMap: {
                    left: "left",
                    top: "bottom",
                    text: ["高", "低"],
                    calculable: true,
                    max: 1100,
                    min: 300,
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
                xAxis: {
                    type: 'value',
                    scale: true,
                    position: 'right',
                    boundaryGap: false,
                    splitLine: {show: false},
                    axisLine: {show: false},
                    axisTick: {show: false},
                    axisLabel: {margin: 2, textStyle: {color: '#aaa'}},
                },
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
