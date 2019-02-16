import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
import { Room, RoomService } from '../services/room.service';
import { StationService, Station } from '../services/station.service';

@Component({
  selector: 'app-home-navigation',
  templateUrl: './home-navigation.component.html',
  styleUrls: ['./home-navigation.component.scss']
})
export class HomeNavigationComponent implements OnInit {

  constructor(private roomService: RoomService,
    private stationService: StationService) { }

  rooms: Room[];
  stations: Station[];
  @Input() section: string;
  @Input() currentItem: string;
  @Output() stationChanged = new EventEmitter<Station>();
  @Output() locationChanged = new EventEmitter<Room>();

  ngOnInit() {
    this.roomService.list()
      .subscribe(res => this.rooms = res);
    this.stationService.list()
      .subscribe(res => this.stations = res.items);
  }

  onStationChanged(value: Station) {
    this.stationChanged.emit(value);
  }

  onLocationChanged(value: Room) {
    this.locationChanged.emit(value);
  }
}
