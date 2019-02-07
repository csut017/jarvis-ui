import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { StationService } from '../services/station.service';

@Component({
  selector: 'app-stations',
  templateUrl: './stations.component.html',
  styleUrls: ['./stations.component.scss']
})
export class StationsComponent implements OnInit {

  constructor(private route: ActivatedRoute,
    private stationService: StationService) { }

  name: string;

  ngOnInit() {
    this.name = this.route.snapshot.paramMap.get('name');
  }

}
