import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { RoomService } from '../services/room.service';

@Component({
  selector: 'app-locations',
  templateUrl: './locations.component.html',
  styleUrls: ['./locations.component.scss']
})
export class LocationsComponent implements OnInit {

  constructor(private route: ActivatedRoute,
    private roomService: RoomService) { }

  name: string;

  ngOnInit() {
    this.name = this.route.snapshot.paramMap.get('name');
  }

}
