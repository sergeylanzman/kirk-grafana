/**
 * Created by taozeyu on 2017/3/10.
 */

import echarts from 'app/plugins/pili/echarts';
import china from './china';
import provinces from './province_in_china';

echarts.registerMap('china', china);

class MapChart {

    constructor(elementId) {
        this._elementId = elementId;
        this._series = [];
        this._seriesMap = {};
        this._legend = {};
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

    setupOption() {
        this._myChart = echarts.init(document.getElementById(this._elementId), 'dark');
        this._legend = {
            show: true,
            orient: "vertical",
            left: "left",
            data: [],
        };
        return {
            legend: this._legend,
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
        };
    }

    updateDataList(option, dataList) {
        this._series.forEach(serie => this._clearSerieData(serie.data));
        dataList.forEach(({targetObj, value}) => {
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

            this._updateSerieData(province, serieName, value);
        });
        this._legend.show = this._legend.data.length > 1;

        option.series = this._series;
        option.legend = this._legend;

        this._myChart.setOption(option);
    }
}

export default MapChart;