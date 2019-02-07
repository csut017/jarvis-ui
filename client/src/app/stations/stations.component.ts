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

  ngOnInit() {
    const name = this.route.snapshot.paramMap.get('name');
    this.loadStation(name);
  }

  loadStation(name: string) {
    this.station = null;
    this.name = name;
    this.stationService.get(this.name)
      .subscribe(res => {
        this.station = res;
      });
  }

}
