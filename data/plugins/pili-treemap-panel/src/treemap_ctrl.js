/**
 * Created by sunfei on 2017/2/23.
 */
import {MetricsPanelCtrl} from 'app/plugins/sdk';
import echarts from 'app/plugins/pili/echarts';

export class TreeMapCtrl extends MetricsPanelCtrl {
    constructor($scope, $injector, $rootScope) {
        super($scope, $injector);
        this.$rootScope = $rootScope;
        let panelDefaults = {
            treeMapScheme:'default',
        };

        _.defaults(this.panel, panelDefaults);
        this.events.on('render', this.onRender.bind(this));
        this.events.on('data-received', this.onDataReceived.bind(this));
        this.events.on('data-error', this.onDataError.bind(this));
        this.events.on('init-edit-mode', this.onInitEditMode.bind(this));

        this.didRenderThisComponent = false;
        // this.events.on('panel-teardown', this.onPanelTeardown.bind(this));
    }

    onInitEditMode() {
        this.addEditorTab('Tree', 'public/plugins/pili-treemap-panel/editor.html', 2);
    }

    onDataError() {
        this.data = [];
        this.render();
    }

    onRender() {
        console.log("onRender111");
        // this.data = this.data;
        // ctrl.data = this.data;
    }

    onDataReceived(dataList) {
        var array = [];
        for (var data of dataList) {
            let str = data.target;
            var name = str.split(': ')[1].split('}')[0];
            var value = data.datapoints[0][0] * 100;
            var treeData = {name: name,
                value: value.toFixed(2)
            };
            array.push(treeData);
        }

        if (!this.didRenderThisComponent) {
            var chart = echarts.init(document.getElementById('main'));
            const option = {
                title : {
                    text: '',
                    subtext: ''
                },
                tooltip : {
                    trigger: 'item',
                    formatter: "{b}: {c}"
                },
                toolbox: {
                    show : false,
                    feature : {
                        mark : {show: true},
                        dataView : {show: true, readOnly: false},
                        restore : {show: true},
                        saveAsImage : {show: true}
                    }
                },
                calculable : false,
                series : [
                    {
                        name:'矩形图',
                        type:'treemap',
                        itemStyle: {
                            normal: {
                                label: {
                                    show: true,
                                    formatter: "{b}"
                                },
                                borderWidth: 1
                            },
                            emphasis: {
                                label: {
                                    show: true
                                }
                            }
                        },
                        data:array,
                    }
                ]
            };
            chart.setOption(option);
        }


    }



    // onPanelTeardown() {
    //     if (this.tree) this.tree.remove();
    // }
}

TreeMapCtrl.templateUrl = 'module.html';
