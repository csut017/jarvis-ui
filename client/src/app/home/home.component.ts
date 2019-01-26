import { Component, OnInit } from '@angular/core';
import { SpeechService } from '../services/speech.service';
import { TimeService } from '../services/time.service';
import { WeatherService } from '../services/weather.service';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {

  constructor(private speech: SpeechService,
    private time: TimeService,
    private weather: WeatherService) { }

  ngOnInit() {
  }

  sayWelcome() {
    const dayPart = this.time.getPartOfDay(),
      date = this.time.getFriendlyDate(),
      timePart = this.time.getFriendlyTime();
    this.speech.say('Good ' + dayPart + ' Craig, today is ' + date + '. The time is ' + timePart).subscribe(_ => { });
  }

  sayWeather() {
    this.weather.getWeather().subscribe(res => {
      this.speech.say(res.current)
        .subscribe(_ => this.speech.say(res.forecast).subscribe(_ => { }));
    });
  }
}
