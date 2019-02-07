import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { StationService, Station } from '../services/station.service';

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
  source: string;

  ngOnInit() {
    const name = this.route.snapshot.paramMap.get('name'),
      source = this.route.snapshot.paramMap.get('source');
    this.loadStation(name);
    if (source) {
      this.source = source;
    }
  }

  loadStation(name: string) {
    this.station = null;
    this.source = null;
    this.name = name;
    this.stationService.get(this.name)
      .subscribe(res => {
        this.station = res;
      });
  }

}
