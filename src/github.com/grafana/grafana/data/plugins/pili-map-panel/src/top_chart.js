/**
 * Created by taozeyu on 2017/3/10.
 */

import echarts from 'app/plugins/pili/echarts';

class TopChart {

    constructor(elementId) {
        this._elementId = elementId;
    }

    setupOption() {
        this._myChart = echarts.init(document.getElementById(this._elementId), 'dark');
        return {
            title: {
                text: '全国主要城市 PM 2.5',
                subtext: 'data from PM25.in',
                sublink: 'http://www.pm25.in',
                left: 'center',
                textStyle: {
                    color: '#fff'
                }
            },
            tooltip : {
                trigger: 'item'
            },
            grid: {
                right: 40,
                top: 100,
                bottom: 40,
                width: '30%'
            },
            xAxis: {
                type: 'value',
                scale: true,
                position: 'top',
                boundaryGap: false,
                splitLine: {show: false},
                axisLine: {show: false},
                axisTick: {show: false},
                axisLabel: {margin: 2, textStyle: {color: '#aaa'}},
            },
            yAxis: {
                type: 'category',
                name: 'TOP 20',
                nameGap: 16,
                axisLine: {show: false, lineStyle: {color: '#ddd'}},
                axisTick: {show: false, lineStyle: {color: '#ddd'}},
                axisLabel: {interval: 0, textStyle: {color: '#ddd'}},
                data: [],
            },
            series : [
                {
                    id: 'bar',
                    zlevel: 2,
                    type: 'bar',
                    symbol: 'none',
                    itemStyle: {
                        normal: {
                            color: '#ddb926'
                        }
                    },
                    data: [],
                },
            ],

        };
    }

    updateDataList(option, dataList) {
        dataList = dataList.sort((element1, element2) => element2.value - element1.value);

        while (dataList.length > 10) {
            dataList.pop();
        }
        option.yAxis = option.yAxis || {data: []};
        option.series = option.series || [{
            id: 'bar', data: []
        }];

        const yAxisData = option.yAxis.data;
        const seriesData = option.series[0].data;

        dataList.reverse().forEach(({targetObj: {province}, value}) => {
            yAxisData.push(`${province} :`);
            seriesData.push(value);
        });

        this._myChart.setOption(option);
    }
}

export default TopChart;