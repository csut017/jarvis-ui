import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { RoomService, Room } from '../services/room.service';

@Component({
  selector: 'app-locations',
  templateUrl: './locations.component.html',
  styleUrls: ['./locations.component.scss']
})
export class LocationsComponent implements OnInit {

  constructor(private route: ActivatedRoute,
    private roomService: RoomService) { }

  name: string;
  location: Room;

  ngOnInit() {
    const name = this.route.snapshot.paramMap.get('name');
    this.loadLocation(name);
  }

  loadLocation(name: string) {
    this.location = null;
    this.name = name;
    this.roomService.get(this.name)
      .subscribe(res => {
        this.location = res;
      });
  }

}
