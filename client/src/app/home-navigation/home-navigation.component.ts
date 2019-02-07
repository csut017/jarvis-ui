import { Component, OnInit, Input } from '@angular/core';
import { RoomInterface, RoomService } from '../services/room.service';
import { StationInterface, StationService } from '../services/station.service';

@Component({
  selector: 'app-home-navigation',
  templateUrl: './home-navigation.component.html',
  styleUrls: ['./home-navigation.component.scss']
})
export class HomeNavigationComponent implements OnInit {

  constructor(private roomService: RoomService,
    private stationService: StationService) { }

  rooms: RoomInterface[];
  stations: StationInterface[];
  @Input() section: string;
  @Input() currentItem: string;

  ngOnInit() {
    this.roomService.list()
      .subscribe(res => this.rooms = res);
    this.stationService.list()
      .subscribe(res => this.stations = res);
  }

}
