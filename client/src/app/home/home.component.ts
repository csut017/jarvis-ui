import { Component, OnInit } from '@angular/core';
import { SpeechService } from '../services/speech.service';
import { TimeService } from '../services/time.service';
import { WeatherService } from '../services/weather.service';
import { RoomService } from '../services/room.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {

  constructor(private speech: SpeechService,
    private time: TimeService,
    private weather: WeatherService,
    private room: RoomService) { }

  ngOnInit() {
  }
}
