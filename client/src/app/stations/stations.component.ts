import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { StationService, Station } from '../services/station.service';
import * as Highcharts from 'highcharts';
import * as moment from 'moment';
import { Source } from '../services/source.service';

@Component({
  selector: 'app-stations',
  templateUrl: './stations.component.html',
  styleUrls: ['./stations.component.scss']
})
export class StationsComponent implements OnInit {

  constructor(private route: ActivatedRoute,
    private stationService: StationService) { }

  name: string;
  station: Station;
  sourceDetails: Source;
  sourceName: string;
  loadError: string;
  effectors: effector[];
  sensors: string[];
  chartInstance: any;

  highcharts = Highcharts;
  chartOptions = {
    chart: {
      type: 'spline'
    },
    series: [],
    time: {
      useUTC: false
    },
    title: {
      text: ''
    },
    tooltip: {
      formatter: function () {
        var s = '<b>' + moment(this.x).format("D MMM YYYY, h:mm a") + '</b>';
        this.points.forEach(point => {
          s += '<br/>' + point.series.name + ': ' +
            point.y + (point.series.tooltipOptions.pointFormat || '');
        });
        return s;
      },
      shared: true
    },
    xAxis: {
      title: {
        text: 'Time'
      },
      type: 'datetime',
      crosshair: true
    }
  };

  ngOnInit() {
    const name = this.route.snapshot.paramMap.get('name'),
      source = this.route.snapshot.paramMap.get('source');
    if (source) {
      this.sourceName = source;
    }
    this.loadStation(name);
  }

  loadStation(name: string):void {
    this.station = null;
    this.name = name;
    this.sourceDetails = null;
    this.stationService.get(this.name)
      .subscribe(res => {
        this.station = res.item;
        this.loadError = res.message;
        if (this.sourceName) {
          this.sourceDetails = this.station.sources.find(s => s.name == this.sourceName);
          this.effectors = this.sourceDetails.effectors.map(e => new effector(e));
          this.loadValues();
          this.sensors = this.sourceDetails.sensors.filter(s => s != 'time');
          this.chartOptions.series = this.sensors.map(s => ({
            name: s,
            data: [],
            marker: {
                radius: 3
            },
            tooltip: {
                pointFormat: ''
            }
          }));
        }
      });
  }

  loadValues(): void {
    this.stationService.getValues(this.name, this.sourceName)
    .subscribe(res => {
      if (res.success) {
        let dataMapping = {},
          chartData = [];
        this.sensors.forEach((s, i) => {
          dataMapping[s] = i;
          chartData[i] = [];
        });
        res.item.items.forEach(s => {
          const time = Date.parse(s.time);
          s.values.forEach(v => {
            const index = dataMapping[v.name];
            if (index || (index === 0)) {
              chartData[index].push([time, v.value]);
            }
          });
        });
        for (let index in chartData) {
          this.chartInstance.series[index].setData(chartData[index], false);
        }
        this.chartInstance.redraw();
      } else {
        this.loadError = 'Unable to retrieve sensor values';
      }
    });
  }

  turnEffectorOn(eff: effector): void {
    this.stationService.sendEffectorCommand(this.name, this.sourceName, eff.name, 'on', eff.duration)
      .subscribe(res => {
        
      });
  }

  turnEffectorOff(eff: effector): void {
    this.stationService.sendEffectorCommand(this.name, this.sourceName, eff.name, 'off', 0)
      .subscribe(res => {
        
      });
  }

  storeChartInstance(chartInstance: any): void {
    this.chartInstance = chartInstance;
  }
}

class effector {
  duration: number = 1;

  constructor(public name: string) {

  }
}