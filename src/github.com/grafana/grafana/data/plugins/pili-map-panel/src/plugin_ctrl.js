import {MetricsPanelCtrl} from 'app/plugins/sdk';
import './dark_theme';
import _ from 'lodash';

import MapChart from './map_chart';
import TopChart from './top_chart';

export class PluginCtrl extends MetricsPanelCtrl {

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

        this._mapChart = new MapChart('map-dom');
        this._topChart = new TopChart('top-dom');
        this._didRenderThisComponent = false;
        this.render();
    }

    setUnitFormat(subItem) {

    }

    changeSeriesColor(series, color) {
        series.color = color;
    }

    onDataReceived(dataList) {
        let mapOption = {},
            topOption = {};

        if (!this._didRenderThisComponent) {
            mapOption = this._mapChart.setupOption();
            topOption = this._topChart.setupOption();
            this._didRenderThisComponent = true;
        }
        this._mapChart.updateDataList(mapOption, dataList);
        this._topChart.updateDataList(topOption, dataList);
    }

    onDataError() {
        console.error(`fail to load database`)
    }

    onRender() {
    }
}

PluginCtrl.templateUrl = 'module.html';
