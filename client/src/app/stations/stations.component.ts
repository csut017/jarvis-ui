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

  highcharts = Highcharts;
  chartOptions = {
    chart: {
      type: 'spline'
    },
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

  loadStation(name: string) {
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
        }
      });
  }

  turnEffectorOn(eff: effector): void {

  }

  turnEffectorOff(eff: effector): void {
    
  }
}

class effector {
  duration: number = 1;

  constructor(public name: string) {

  }
}